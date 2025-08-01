package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kc2g-flex-tools/flexclient"
)

func init() {
	hamlib.AddHandler(
		NewHandler(
			"get_mode", "m",
			get_mode,
			Args(0),
			FieldNames("Mode", "Passband"),
		),
	)

	hamlib.AddHandler(
		NewHandler(
			"set_mode", "M",
			set_mode,
			Args(2),
		),
	)
}

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

var modeReversed = map[string]bool{
	"LSB":  true,
	"DIGL": true,
}

var defaultWidth = map[string]int{
	"AM":   6000,
	"SAM":  6000,
	"USB":  2700,
	"LSB":  2700,
	"CW":   500,
	"DIGU": 3000,
	"DIGL": 3000,
	"FM":   0, // not settable
	"DFM":  12000,
}

var centerFreq = map[string]int{
	// two-sided modes
	"AM":  0,
	"SAM": 0,
	"FM":  0,
	"DFM": 0,
	// USB
	"USB":  1500,
	"DIGU": 1500,
	// LSB
	"LSB":  -1500,
	"DIGL": -1500,
}

func get_mode(ctx context.Context, _ []string) (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("get slice %s failed", SliceIdx)
	}

	translated, ok := modesFromFlex[slice["mode"]]
	if !ok {
		return "", fmt.Errorf("unknown mode %s", slice["mode"])
	}

	lo, err := strconv.Atoi(slice["filter_lo"])
	if err != nil {
		return "", err
	}
	hi, err := strconv.Atoi(slice["filter_hi"])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n%d\n", translated, hi-lo), nil
}

func set_mode(ctx context.Context, args []string) (string, error) {
	mode, ok := modesToFlex[args[0]]
	if !ok {
		return "", fmt.Errorf("unknown mode %s", args[0])
	}

	width, err := strconv.Atoi(args[1])
	if err != nil {
		return "", err
	}

	res, err := fc.SliceSet(ctx, SliceIdx, flexclient.Object{"mode": mode})
	if err != nil {
		return "", err
	}
	if res.Error != 0 && res.Error != 0x50002001 {
		return "", fmt.Errorf("slice set %08X", res.Error)
	}

	if width < 0 {
		width = defaultWidth[mode]
	}

	if width != 0 {
		var lo, hi int
		lo = centerFreq[mode] - width/2
		hi = centerFreq[mode] + width/2

		res, err := fc.SliceSetFilter(ctx, SliceIdx, lo, hi)
		if err != nil {
			return "", err
		}
		if res.Error != 0 {
			return "", fmt.Errorf("set filter %08X", res.Error)
		}
	}

	return Success, nil
}
