package main

import (
	"fmt"
	"strconv"

	"github.com/arodland/flexclient"
)

func init() {
	hamlib.AddHandler(
		names{{`l`, `AF`}, {`\get_level`, `AF`}},
		NewHandler(
			get_level_af,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`L`, `AF`}, {`\set_level`, `AF`}},
		NewHandler(
			set_level_af,
			Args(1),
		),
	)
}

func get_level_af(_ Conn, _ []string) (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}

	audio_level, err := strconv.ParseFloat(slice["audio_level"], 64)
	if err != nil {
		return "", err
	}
	audio_level /= 100
	return fmt.Sprintf("%.3f\n", audio_level), nil
}

func set_level_af(_ Conn, args []string) (string, error) {
	audio_level, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return "", err
	}
	audio_level *= 100

	val := fmt.Sprintf("%.0f", audio_level)
	res := fc.SliceSet(SliceIdx, flexclient.Object{"audio_level": val})
	if res.Error != 0 {
		return "", fmt.Errorf("slice set %08X", res.Error)
	}

	return Success, nil
}
