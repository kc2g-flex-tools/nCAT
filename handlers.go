package main

import "github.com/arodland/flexclient"

import (
	"fmt"
	"strconv"
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
	hamlib.AddHandler(`\dump_state`, func(_ []string) string {
		return `0
2
1
0.000000 10000000000.000000 0xef -1 -1 0x1 0x0
0 0 0 0 0 0 0
0 0 0 0 0 0 0
0xef 1
0xef 0
0 0
0x82 500
0x82 200
0x82 2000
0x21 10000
0x21 5000
0x21 20000
0x0c 2700
0x0c 1400
0x0c 3900
0x40 160000
0x40 120000
0x40 200000
0 0
0
0
0
0
0
0
0
0x40000000
0x40000020
0x20
0
0
`
	})
	hamlib.AddHandler("v", func(_ []string) string {
		return "VFOA\n"
	})
	hamlib.AddHandler("V", func(args []string) string {
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
	})
	hamlib.AddHandler("m", func(_ []string) string {
		slice, ok := fc.GetObject("slice " + SliceIdx)
		if !ok {
			return "ERR\n0\n"
		}

		translated, ok := modesFromFlex[slice["mode"]]
		if !ok {
			return "ERR\n0\n"
		}
		return translated + "\n3000\n"
	})
	hamlib.AddHandler("M", func(args []string) string {
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

		var update flexclient.Object

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
	})
	hamlib.AddHandler("f", func(_ []string) string {
		slice, ok := fc.GetObject("slice " + SliceIdx)
		if !ok {
			return "ERR\n"
		}

		freq, err := strconv.ParseFloat(slice["RF_frequency"], 64)
		if err != nil {
			return "ERR\n"
		}
		return fmt.Sprintf("%f\n", freq*1e6)
	})
	hamlib.AddHandler("F", func(args []string) string {
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
	})
	hamlib.AddHandler("U", func(args []string) string {
		if len(args) == 2 && args[0] == "TUNER" {
			res := fc.SendAndWait("transmit tune " + args[1])
			if res.Error == 0 {
				return "RPRT 0\n"
			} else {
				return "RPRT 1\n"
			}
		} else {
			return "RPRT 1\n"
		}
	})
	hamlib.AddHandler("T", func(args []string) string {
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
	})
}
