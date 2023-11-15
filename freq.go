package main

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/kc2g-flex-tools/flexclient"
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
		names{{`j`}, {`\get_rit`}},
		NewHandler(
			get_ritxit("rit"),
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`J`}, {`\set_rit`}},
		NewHandler(
			set_ritxit("rit"),
			Args(1),
		),
	)

	hamlib.AddHandler(
		names{{`z`}, {`\get_xit`}},
		NewHandler(
			get_ritxit("xit"),
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`Z`}, {`\set_xit`}},
		NewHandler(
			set_ritxit("xit"),
			Args(1),
		),
	)
}

func get_freq(ctx context.Context, _ []string) (string, error) {
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

func set_freq(ctx context.Context, args []string) (string, error) {
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

func get_ritxit(which string) func(context.Context, []string) (string, error) {
	return func(ctx context.Context, _ []string) (string, error) {
		slice, ok := fc.GetObject("slice " + SliceIdx)

		if !ok {
			return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
		}

		on, err := strconv.Atoi(slice[which+"_on"])
		if err != nil {
			return "", err
		}

		freq, err := strconv.Atoi(slice[which+"_freq"])
		if err != nil {
			return "", err
		}

		if on == 0 {
			freq = 0
		}

		return fmt.Sprintf("%d\n", freq), nil
	}
}

func set_ritxit(which string) func(context.Context, []string) (string, error) {
	return func(ctx context.Context, args []string) (string, error) {
		freq, err := strconv.Atoi(args[0])
		if err != nil {
			return "", err
		}

		obj := flexclient.Object{}
		if freq == 0 {
			obj[which+"_on"] = "0"
			obj[which+"_freq"] = "0"
		} else {
			obj[which+"_on"] = "1"
			obj[which+"_freq"] = fmt.Sprintf("%d", freq)
		}

		res := fc.SliceSet(SliceIdx, obj)
		if res.Error != 0 {
			return "", fmt.Errorf("slice set %08X", res.Error)
		}

		return Success, nil
	}
}
