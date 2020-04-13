package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/arodland/flexclient"
)

var modesToFlex = map[string]string{
	"AM":     "AM",
	"AMS":    "SAM",
	"USB":    "USB",
	"LSB":    "LSB",
	"CW":     "CW",
	"PKTUSB": "DIGU",
	"PKTLSB": "DIGL",
	"FM":     "FM",
	"PKTFM":  "DFM",
}

var modesFromFlex = map[string]string{
	"AM":   "AM",
	"SAM":  "AMS",
	"USB":  "USB",
	"LSB":  "LSB",
	"CW":   "CW",
	"DIGU": "PKTUSB",
	"DIGL": "PKTLSB",
	"FM":   "FM",
	"DFM":  "PKTFM",
}

func RegisterHandlers() {
	hamlib.AddHandler(func(_ []string) string {
		return "0\n" + // protocol version
			"2\n" + // hamlib model
			"2\n" + // region
			"30000.000000 54000000.000000 0xe2f -1 -1 0x1 0x0\n" + // RX: 30kHz - 54MHz, AM|CW|USB|LSB|FM|AMS|PKTUSB|PKTLSB
			"0 0 0 0 0 0 0\n" + // end of RX range list
			"100000.000000 54000000.000000 0xe2f 1000 100000 0x1 0x0\n" + // TX: 100kHz - 54MHz, 1-100 watts, AM|CW|USB|LSB|FM|AMS|PKTUSB|PKTLSB
			"0 0 0 0 0 0 0\n" + // end of TX range list
			"0xe2f 1\n" +
			"0xe2f 0\n" +
			"0 0\n" + // end of tuning steps
			"0x02 500\n" + // CW normal
			"0x02 200\n" + // CW narrow
			"0x02 2000\n" + // CW wide
			"0x02 50\n" + // CW min
			"0x02 10000\n" + // CW max
			"0x02 0\n" + // allow arbitrary
			"0x221 6000\n" + // AM|FM|AMS normal
			"0x221 3000\n" + // AM|FM|AMS narrow
			"0x221 10000\n" + // AM|FM|AMS wide
			"0x221 100\n" + // AM|FM|AMS min
			"0x221 20000\n" + // AM|FM|AMS max
			"0x221 0\n" + // allow arbitrary
			"0x0c 2700\n" + // SSB normal
			"0x0c 1400\n" + // SSB narrow
			"0x0c 3900\n" + // SSB wide
			"0x0c 50\n" + // SSB min
			"0x0c 10000\n" + // SSB max
			"0x0c 0\n" + // allow arbitrary
			"0xc00 3000\n" + // digi normal
			"0xc00 1500\n" + // digi narrow
			"0xc00 4000\n" + // digi wide
			"0xc00 50\n" + // digi min
			"0xc00 10000\n" + //digi max
			"0xc00 0\n" + // allow arbitrary
			"0 0\n" + // end of filter widths
			"0\n" + // max rit
			"0\n" + // max xit
			"0\n" + // max if_shift
			"0\n" + // no announce capabilities
			"0 8 16 24 32\n" + // preamp
			"0 8\n" + // attenuator
			"0x48400833be\n" + // func get: NB|COMP|VOX|TONE|TSQL|FBKIN|ANF|NR|MON|MN|REV|TUNER|ANL|DIVERSITY
			"0x48400833be\n" + // func set: NB|COMP|VOX|TONE|TSQL|FBKIN|ANF|NR|MON|MN|REV|TUNER|ANL|DIVERSITY
			"0x600023710f\n" + // level get: PREAMP|ATT|VOXDELAY|NR|RFPOWER|MICGAIN|KEYSPD|COMP|AGC|VOXGAIN|MONITOR_GAIN|NB (TODO: use metering protocol to add SWR|ALC|RFPOWER_METER|COMP_METER)
			"0x600023710f\n" + // level set: PREAMP|ATT|VOXDELAY|NR|RFPOWER|MICGAIN|KEYSPD|COMP|AGC|VOXGAIN|MONITOR_GAIN|NB
			"0\n" + // parm get: none
			"0\n" // parm set: none
	}, `\dump_state`)
	hamlib.AddHandler(func(_ []string) string {
		return "VFOA\n"
	}, "v", `\get_vfo`)
	hamlib.AddHandler(func(args []string) string {
		if len(args) != 1 {
			return "RPRT 1\n"
		}

		if args[0] == "?" {
			return "VFOA\n"
		} else if args[0] == "VFOA" {
			return "RPRT 0\n"
		} else {
			return "RPRT 1\n"
		}
	}, "V", `\set_vfo`)
	hamlib.AddHandler(func(_ []string) string {
		slice, ok := fc.GetObject("slice " + SliceIdx)
		if !ok {
			return "ERR\n0\n"
		}

		translated, ok := modesFromFlex[slice["mode"]]
		if !ok {
			return "ERR\n0\n"
		}
		return translated + "\n3000\n"
	}, "m", `\get_mode`)
	hamlib.AddHandler(func(args []string) string {
		if len(args) != 2 {
			return "RPRT 1\n"
		}
		mode, ok := modesToFlex[args[0]]
		if !ok {
			return "RPRT 1\n"
		}

		width, err := strconv.Atoi(args[1])
		if err != nil {
			return "RPRT 1\n"
		}

		if width < 0 || width > 3000 {
			width = 3000
		}

		update := flexclient.Object{}

		update["mode"] = mode

		var lo, hi int
		if width != 0 {
			lo = 1500 - (width / 2)
			hi = 1500 + (width / 2)

			update["filter_lo"] = fmt.Sprintf("%d", lo)
			update["filter_hi"] = fmt.Sprintf("%d", hi)
		}

		res := fc.SliceSet(SliceIdx, update)

		if res.Error == 0 {
			return "RPRT 0\n"
		} else {
			fmt.Printf("%#v\n", res)
			return "RPRT 1\n"
		}
	}, "M", `\set_mode`)
	hamlib.AddHandler(func(_ []string) string {
		slice, ok := fc.GetObject("slice " + SliceIdx)
		if !ok {
			return "ERR\n"
		}

		freq, err := strconv.ParseFloat(slice["RF_frequency"], 64)
		if err != nil {
			return "ERR\n"
		}
		return fmt.Sprintf("%d\n", int64(math.Round(freq*1e6)))
	}, "f", `\get_freq`)
	hamlib.AddHandler(func(args []string) string {
		if len(args) != 1 {
			return "RPRT 1\n"
		}
		freq, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return "RPRT 1\n"
		}

		res := fc.SliceTune(SliceIdx, freq/1e6)

		if res.Error == 0 {
			return "RPRT 0\n"
		} else {
			fmt.Printf("%#v\n", res)
			return "RPRT 1\n"
		}
	}, "F", `\set_freq`)
	hamlib.AddHandler(func(args []string) string {
		if len(args) == 2 && args[0] == "TUNER" {
			res := fc.TransmitTune(args[1])
			if res.Error == 0 {
				return "RPRT 0\n"
			}
		}
		return "RPRT 1\n"
	}, "U", `\set_func`)
	hamlib.AddHandler(func(args []string) string {
		if len(args) != 1 {
			return "RPRT 1\n"
		}

		if args[0] == "TUNER" {
			xmit, ok := fc.GetObject("transmit")
			if !ok {
				return "RPRT 1\n"
			}
			return xmit["tune"] + "\n"
		}
		return "RPRT 1\n"
	}, "u", `\get_func`)
	hamlib.AddHandler(func(args []string) string {
		if len(args) == 1 {
			tx := "1"
			if args[0] == "0" {
				tx = "0"
			}
			res := fc.SendAndWait("xmit " + tx)
			if res.Error == 0 {
				return "RPRT 0\n"
			}
		}
		return "RPRT 1\n"
	}, "T", `\set_ptt`)
	hamlib.AddHandler(func(_ []string) string {
		interlock, ok := fc.GetObject("interlock")
		if !ok {
			return "RPRT 1\n"
		}
		if interlock["state"] == "TRANSMITTING" {
			return "1\n"
		} else {
			return "0\n"
		}
	}, "t", `\get_ptt`)
	hamlib.AddHandler(func(args []string) string {
		if len(args) == 2 && args[0] == "RFPOWER" {
			power, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return "RPRT 1\n"
			}
			power *= 100
			res := fc.TransmitSet(flexclient.Object{"rfpower": fmt.Sprintf("%.0f", power)})
			if res.Error == 0 {
				return "RPRT 0\n"
			}
		}
		return "RPRT 1\n"
	}, "L", `\set_level`)
	hamlib.AddHandler(func(args []string) string {
		if len(args) != 1 {
			return "RPRT 1\n"
		}

		if args[0] == "RFPOWER" {
			xmit, ok := fc.GetObject("transmit")
			if !ok {
				return "RPRT 1\n"
			}
			power, err := strconv.ParseFloat(xmit["rfpower"], 64)
			if err != nil {
				return "RPRT 1\n"
			}
			power /= 100
			return fmt.Sprintf("%.3f\n", power)
		}
		return "RPRT 1\n"
	}, "l", `\get_level`)
}
