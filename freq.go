package main

import (
	"fmt"
	"math"
	"strconv"
)

func init() {
	hamlib.AddHandler(
		names{{`f`}, {`\get_freq`}},
		NewHandler(
			get_freq,
			Args(0),
			ErrResponse("ERR\n"),
		),
	)

	hamlib.AddHandler(
		names{{`F`}, {`\set_freq`}},
		NewHandler(
			set_freq,
			Args(1),
		),
	)
}

func get_freq(_ Conn, _ []string) (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}

	freq, err := strconv.ParseFloat(slice["RF_frequency"], 64)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d\n", int64(math.Round(freq*1e6))), nil
}

func set_freq(_ Conn, args []string) (string, error) {
	freq, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return "", err
	}

	res := fc.SliceTune(SliceIdx, freq/1e6)

	if res.Error != 0 {
		return "", fmt.Errorf("slice tune %08X", res.Error)
	}

	return Success, nil
}
