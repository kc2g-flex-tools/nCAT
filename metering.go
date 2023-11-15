package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/kc2g-flex-tools/flexclient"
	log "github.com/rs/zerolog/log"
)

type conversion func(float64) float64

var meters = map[string]conversion{
	"SWR":                 nil,
	"ALC":                 nil,
	"STRENGTH":            func(dBm float64) float64 { return dBm + 73 },
	"RFPOWER_METER":       dBmToRFPower,
	"COMP_METER":          nil,
	"VD_METER":            nil,
	"RFPOWER_METER_WATTS": dBmToWatts,
	"TEMP_METER":          nil,
}

var meterMu sync.RWMutex
var meterVal = map[string]float64{}
var meterMeta = map[string]flexclient.Object{}
var flexToHamlib = map[string]string{}
var hamlibToFlex = map[string]string{}
var pktReader bytes.Reader

func init() {
	for meter := range meters {
		handler := NewHandler("get_level", "l",
			get_level_metering,
			RequiredArgs(meter),
			AllArgs(true),
			Args(2),
		)

		hamlib.AddHandler(handler)
	}
}

func enableMetering(fc *flexclient.FlexClient) {
	if cfg.UDPPort != 0 {
		fc.SetUDPPort(cfg.UDPPort)
	}
	err := fc.InitUDP()
	if err != nil {
		log.Fatal().Err(err).Msg("fc.initUDP failed")
	}
	go fc.RunUDP()

	meterPackets := make(chan flexclient.VitaPacket, 50)
	fc.SetVitaChan(meterPackets)

	meterEvents := make(chan flexclient.StateUpdate)
	fc.Subscribe(flexclient.Subscription{"meter ", meterEvents})
	go handleMeters(fc, meterPackets, meterEvents)
	fc.SendAndWait("sub meter all")
}

func handleMeters(fc *flexclient.FlexClient, meterPackets chan flexclient.VitaPacket, meterEvents chan flexclient.StateUpdate) {
	for {
		select {
		case upd, ok := <-meterEvents:
			if !ok {
				log.Info().Msg("handle_meters exiting on update chan closed")
				return
			}
			meterUpdate(upd)
		case pkt, ok := <-meterPackets:
			if !ok {
				log.Info().Msg("handle_meters exiting on meter chan closed")
				return
			}
			meterPacket(pkt)
		}
	}
}

func meterUpdate(upd flexclient.StateUpdate) {
	meterMu.Lock()
	defer meterMu.Unlock()
	meterName, meter := upd.Object, upd.CurrentState
	log.Trace().Interface("meter_name", meterName).Interface("metadata", meter).Send()

	if len(meter) == 0 { // removal
		rev, found := flexToHamlib[meterName]
		if found {
			delete(flexToHamlib, meterName)
			delete(hamlibToFlex, rev)
			meterVal[meterName] = 0
			log.Debug().Str("name", meterName).Str("hamlib", rev).Msg("unmapped meter")
		}
		return
	}

	var dest string
	switch {
	case meter["src"] == "TX-" && meter["nam"] == "HWALC":
		dest = "ALC"
	case meter["src"] == "TX-" && meter["nam"] == "SWR":
		dest = "SWR"
	case meter["src"] == "TX-" && meter["nam"] == "COMPPEAK":
		dest = "COMP_METER"
	case meter["src"] == "TX-" && meter["nam"] == "FWDPWR":
		dest = "RFPOWER_METER_WATTS"
	case meter["src"] == "TX-" && meter["nam"] == "PATEMP":
		dest = "TEMP_METER"
	case meter["src"] == "RAD" && meter["nam"] == "+13.8A":
		dest = "VD_METER"
	case meter["src"] == "SLC" && meter["num"] == SliceIdx && meter["nam"] == "LEVEL":
		dest = "STRENGTH"
	}
	if dest != "" {
		meterMeta[meterName] = meter
		if oldH2F, found := hamlibToFlex[dest]; found {
			delete(flexToHamlib, oldH2F)
		}
		flexToHamlib[meterName] = dest
		hamlibToFlex[dest] = meterName
		log.Debug().Str("name", meterName).Str("hamlib", dest).Msg("mapped meter")
	}
}

func meterPacket(pkt flexclient.VitaPacket) {
	meterMu.Lock()
	defer meterMu.Unlock()
	log.Trace().Hex("payload", pkt.Payload).Msg("meter packet")
	pktReader.Reset(pkt.Payload)
	for {
		var id uint16
		var rawVal int16
		var val float64

		err := binary.Read(&pktReader, binary.BigEndian, &id)
		if err == io.EOF {
			return
		} else if err != nil {
			log.Fatal().Err(err).Msg("decoding meter packet (1)")
		}
		name := fmt.Sprintf("meter %d", id)
		hamlib, found := flexToHamlib[name]
		if !found {
			continue
		}
		meta, found := meterMeta[name]
		if !found {
			continue
		}
		err = binary.Read(&pktReader, binary.BigEndian, &rawVal)
		if err != nil {
			log.Fatal().Err(err).Msg("decoding meter packet (2)")
		}

		switch meta["unit"] {
		case "dBm", "dBFS", "SWR":
			val = float64(rawVal) / 128
		case "Volts":
			val = float64(rawVal) / 256
		case "degF", "degC":
			val = float64(rawVal) / 64
		default:
			val = float64(rawVal)
			log.Debug().Msgf("don't know how to decode a meter packet with unit %s, passing through", meta["unit"])
		}

		log.Trace().Int("id", int(id)).Int("rawval", int(rawVal)).Float64("val", val).Str("hamlib", hamlib).Interface("meta", meta).Msg("meter data")
		meterVal[name] = val
	}
}

func dBmToWatts(dBm float64) float64 {
	if dBm <= 0 {
		// The meter rests at 0 when not transmitting, and real fwd power is always going to be positive.
		// Easier than reading the transmitter state.
		return 0
	}
	return math.Pow(10, (dBm/10)-3)
}

func dBmToRFPower(dBm float64) float64 {
	// As far as I know, this doesn't read at all for transverter/2m ports (neither does SWR).
	// You might think that the max power would be in "info radio" somewhere but it's not.
	return dBmToWatts(dBm) / 100
}

func get_level_metering(ctx context.Context, args []string) (string, error) {
	level := args[1]
	conv, ok := meters[level]
	if !ok {
		log.Warn().Str("hamlib_level", level).Msg("get_level_metering: level was never registered, shouldn't happen")
		return Error, nil
	}
	if level == "RFPOWER_METER" {
		level = "RFPOWER_METER_WATTS" // same except for the conversion
	}
	meterMu.RLock()
	defer meterMu.RUnlock()

	meterName, ok := hamlibToFlex[level]
	if !ok {
		log.Warn().Str("hamlib_level", level).Msg("get_level_metering: but no metadata heard from radio")
		return Error, nil
	}
	val, ok := meterVal[meterName]
	if !ok {
		log.Warn().Str("hamlib_level", level).Str("meter_name", meterName).Msg("get_level_metering: no meter value received")
	}

	if conv != nil {
		val = conv(val)
	}

	return fmt.Sprintf("%.3f\n", val), nil
}
