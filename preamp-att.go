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
		names{{`l`, `PREAMP`}, {`\get_level`, `PREAMP`}, {`l`, `ATT`}, {`\get_level`, `ATT`}},
		NewHandler(
			get_level_preamp_att,
			AllArgs(true),
			Args(2),
		),
	)

	hamlib.AddHandler(
		names{{`L`, `PREAMP`}, {`\set_level`, `PREAMP`}, {`L`, `ATT`}, {`\set_level`, `ATT`}},
		NewHandler(
			set_level_preamp_att,
			AllArgs(true),
			Args(3),
		),
	)
}

func get_level_preamp_att(ctx context.Context, args []string) (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}

	gain, err := strconv.Atoi(slice["rfgain"])
	if err != nil {
		return "", err
	}

	if args[1] == "ATT" {
		if gain < 0 {
			return "1\n", nil
		} else {
			return "0\n", nil
		}
	} else { // PREAMP
		if gain < 0 {
			return "0\n", nil
		} else { // 0, 1, 2, 3, 4
			return fmt.Sprintf("%d\n", gain/8), nil
		}
	}
}

func set_level_preamp_att(ctx context.Context, args []string) (string, error) {
	level, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return "", err
	}

	level = 8 * math.Round(level)

	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}

	currLevel, err := strconv.ParseFloat(slice["rfgain"], 64)
	if err != nil {
		return "", err
	}

	obj := flexclient.Object{}

	if args[1] == "ATT" {
		// PREAMP and ATT are separate controls to hamlib, but coupled here.
		// Clients might do something like "L PREAMP 1" followed by "L ATT 0".
		// If they do something like that, leave the state alone instead of
		// setting preamp to 0.
		if currLevel >= 0 && level == 0 {
			return Success, nil
		}

		if level == 0 {
			obj["rfgain"] = "0"
		} else {
			obj["rfgain"] = "-8"
		}
	} else {
		// See note above.
		if currLevel <= 0 && level == 0 {
			return Success, nil
		}

		if level < 0 || level > 24 {
			return "", fmt.Errorf("invalid rfgain %f", level)
		}
		obj["rfgain"] = fmt.Sprintf("%.0f", level)
	}

	res := fc.SliceSet(SliceIdx, obj)
	if res.Error != 0 {
		return "", fmt.Errorf("slice set %08X", res.Error)
	}

	return Success, nil
}
