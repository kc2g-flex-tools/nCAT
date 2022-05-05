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

	hamlib.AddHandler(
		names{{`i`}, {`\get_split_freq`}},
		NewHandler(
			get_split_freq,
			Args(0),
		),
	)
	hamlib.AddHandler(
		names{{`I`}, {`\set_split_freq`}},
		NewHandler(
			set_split_freq,
			Args(1),
		),
	)
}

func get_freq(_ Conn, _ []string) (string, error) {
	if swapVFO {
		return get_split_freq_real()
	} else {
		return get_freq_real()
	}
}

func get_freq_real() (string, error) {
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
	if swapVFO {
		return set_split_freq_real(args)
	} else {
		return set_freq_real(args)
	}
}

func set_freq_real(args []string) (string, error) {
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

func get_split_freq(_ Conn, _ []string) (string, error) {
	if swapVFO {
		return get_freq_real()
	} else {
		return get_split_freq_real()
	}
}

func get_split_freq_real() (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}

	rxf, err := strconv.ParseFloat(slice["RF_frequency"], 64)
	if err != nil {
		return "", err
	}

	xit, err := strconv.Atoi(slice["xit_freq"])
	if err != nil {
		return "", err
	}

	freq := int64(math.Round(rxf*1e6) + float64(xit))
	return fmt.Sprintf("%d\n", freq), nil
}

func set_split_freq(_ Conn, args []string) (string, error) {
	if swapVFO {
		return set_freq_real(args)
	} else {
		return set_split_freq_real(args)
	}
}

func set_split_freq_real(args []string) (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}

	rxf, err := strconv.ParseFloat(slice["RF_frequency"], 64)
	if err != nil {
		return "", err
	}

	freq, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return "", err
	}

	offset := int64(math.Round(freq - rxf*1e6))
	return set_ritxit_freq("xit", offset)
}
