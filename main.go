package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/kc2g-flex-tools/flexclient"
	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"
)

var cfg struct {
	RadioIP          string
	Station          string
	Slice            string
	Headless         bool
	SliceCreateParms string
	Listen           string
	Profile          string
	LogLevel         string
}

func init() {
	flag.StringVar(&cfg.RadioIP, "radio", ":discover:", "radio IP address or discovery spec")
	flag.StringVar(&cfg.Station, "station", "Flex", "station name to bind to or create")
	flag.StringVar(&cfg.Slice, "slice", "A", "slice letter to control")
	flag.BoolVar(&cfg.Headless, "headless", false, "run in headless mode")
	flag.StringVar(&cfg.Listen, "listen", ":4532", "hamlib listen [address]:port")
	flag.StringVar(&cfg.Profile, "profile", "", "global profile to load on startup for -headless mode")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "minimum level of messages to log to console")
}

var fc *flexclient.FlexClient
var hamlib *HamlibServer = NewHamlibServer()
var ClientID string
var ClientUUID string
var SliceIdx string

func createClient() {
	log.Info().Msg("Registering client")
	res := fc.SendAndWait("client gui")
	if res.Error != 0 {
		panic(res)
	}
	ClientUUID = res.Message
	ClientID = "0x" + fc.ClientID()

	fc.SendAndWait("client program Hamlib-Flex")
	fc.SendAndWait("client station " + cfg.Station)

	log.Info().Str("handle", ClientID).Msg("Got client handle")

	if cfg.Profile != "" {
		res := fc.SendAndWait("profile global load " + cfg.Profile)
		if res.Error != 0 {
			log.Printf("Profile load failed: %08X (typo?)", res.Error)
		} else {
			log.Printf("Loaded profile %s", cfg.Profile)
		}
	}
}

func bindClient() {
	log.Info().Str("station", cfg.Station).Msg("Waiting for station")

	clients := make(chan flexclient.StateUpdate)
	sub := fc.Subscribe(flexclient.Subscription{"client ", clients})
	cmdResult := fc.SendNotify("sub client all")

	var found, cmdComplete bool

	for !found || !cmdComplete {
		select {
		case upd := <-clients:
			if upd.CurrentState["station"] == cfg.Station {
				ClientID = strings.TrimPrefix(upd.Object, "client ")
				ClientUUID = upd.CurrentState["client_id"]
				found = true
			}
		case <-cmdResult:
			cmdComplete = true
		}
	}

	fc.Unsubscribe(sub)

	log.Info().Str("client_id", ClientID).Str("uuid", ClientUUID).Msg("Found client")

	fc.SendAndWait("client bind client_id=" + ClientUUID)
}

func findSlice() {
	log.Info().Str("slice_id", cfg.Slice).Msg("Looking for slice")
	slices := make(chan flexclient.StateUpdate)
	sub := fc.Subscribe(flexclient.Subscription{"slice ", slices})
	cmdResult := fc.SendNotify("sub slice all")

	var found, cmdComplete bool

	for !found || !cmdComplete {
		select {
		case upd := <-slices:
			if upd.CurrentState["index_letter"] == cfg.Slice && upd.CurrentState["client_handle"] == ClientID {
				SliceIdx = strings.TrimPrefix(upd.Object, "slice ")
				found = true
			}
		case <-cmdResult:
			cmdComplete = true
		}
	}

	fc.Unsubscribe(sub)
	log.Info().Str("slice_idx", SliceIdx).Msg("Found slice")
}

func main() {
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{
			Out: os.Stderr,
		},
	).With().Timestamp().Logger()

	flag.Parse()

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Str("level", cfg.LogLevel).Msg("Unknown log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	if cfg.Profile != "" && !cfg.Headless {
		log.Fatal().Msg("-profile doesn't make sense without -headless")
	}

	fc, err = flexclient.NewFlexClient(cfg.RadioIP)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		fc.Run()
		hamlib.Close()
		wg.Done()
	}()

	err = hamlib.Listen(cfg.Listen)
	if err != nil {
		panic(err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Info().Msg("Exit on SIGINT")
		fc.Close()
		hamlib.Close()
	}()

	if cfg.Headless {
		createClient()
	} else {
		bindClient()
	}
	findSlice()

	fc.SendAndWait("sub radio all")
	fc.SendAndWait("sub tx all")

	wg.Add(1)
	go func() {
		hamlib.Run()
		fc.Close()
		wg.Done()
	}()

	wg.Wait()
}
