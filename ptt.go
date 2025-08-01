package main

import (
	"context"
	"fmt"

	"github.com/kc2g-flex-tools/flexclient"
)

func init() {
	hamlib.AddHandler(
		NewHandler(
			"get_ptt", "t",
			get_ptt,
			Args(0),
			FieldNames("PTT"),
		),
	)

	hamlib.AddHandler(
		NewHandler(
			"set_ptt", "T",
			set_ptt,
			Args(1),
		),
	)
}

func get_ptt(ctx context.Context, _ []string) (string, error) {
	interlock, ok := fc.GetObject("interlock")
	if !ok {
		return "", fmt.Errorf("couldn't get interlock")
	}

	// TODO: what should this return if a different slice is transmitting?

	if interlock["state"] == "TRANSMITTING" {
		return "1\n", nil
	} else {
		return "0\n", nil
	}
}

func set_ptt(ctx context.Context, args []string) (string, error) {
	tx := "1"
	if args[0] == "0" {
		tx = "0"
	}

	// If requesting PTT on, and the current slice isn't the TX slice,
	// take over the TX flag.
	if tx == "1" {
		slice, ok := fc.GetObject("slice " + SliceIdx)
		if !ok {
			return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
		}
		if slice["tx"] != "1" {
			res, err := fc.SliceSet(ctx, SliceIdx, flexclient.Object{"tx": "1"})
			if err != nil {
				return "", err
			}
			if res.Error != 0 {
				return "", fmt.Errorf("slice set %08X", res.Error)
			}
		}
	}

	res := fc.SendAndWait("xmit " + tx)
	if res.Error != 0 {
		return "", fmt.Errorf("xmit %08X", res.Error)
	}
	return Success, nil
}
