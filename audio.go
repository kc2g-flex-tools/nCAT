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
			"get_level", "l",
			get_level_af,
			RequiredArgs("AF"),
			Args(0),
			FieldNames("Level Value"),
		),
	)

	hamlib.AddHandler(
		NewHandler(
			"set_level", "L",
			set_level_af,
			RequiredArgs("AF"),
			Args(1),
		),
	)
}

func get_level_af(ctx context.Context, _ []string) (string, error) {
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

func set_level_af(ctx context.Context, args []string) (string, error) {
	audio_level, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return "", err
	}
	audio_level *= 100

	val := fmt.Sprintf("%.0f", audio_level)
	res, err := fc.SliceSet(ctx, SliceIdx, flexclient.Object{"audio_level": val})
	if err != nil {
		return "", err
	}
	if res.Error != 0 {
		return "", fmt.Errorf("slice set %08X", res.Error)
	}

	return Success, nil
}
