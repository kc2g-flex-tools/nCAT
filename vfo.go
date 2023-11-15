package main

import (
	"context"
	"fmt"
)

func init() {
	hamlib.AddHandler(
		NewHandler(
			"chk_vfo", "",
			chk_vfo,
			Args(0),
		),
	)
	hamlib.AddHandler(
		NewHandler(
			"get_vfo", "v",
			get_vfo,
			Args(0),
		),
	)
	hamlib.AddHandler(
		NewHandler(
			"set_vfo", "V",
			set_vfo,
			Args(1),
		),
	)
	hamlib.AddHandler(
		NewHandler(
			"get_split_vfo", "s",
			get_split_vfo,
			Args(0),
		),
	)
	hamlib.AddHandler(
		NewHandler(
			"set_split_vfo", "S",
			set_split_vfo,
			Args(2),
		),
	)
	hamlib.AddHandler(
		NewHandler(
			"get_lock_mode", "",
			zero,
			Args(0),
		),
	)
}

func chk_vfo(ctx context.Context, _ []string) (string, error) {
	getConn(ctx).chkVFOexecuted = true
	if cfg.ChkVFOMode == "new" {
		return "0\n", nil
	} else {
		return "CHKVFO 0\n", nil
	}
}

func get_vfo(ctx context.Context, _ []string) (string, error) {
	return "VFOA\n", nil
}

func set_vfo(ctx context.Context, args []string) (string, error) {
	switch args[0] {
	case "?":
		// List available VFOs
		return "VFOA\n", nil
	case "VFOA", "Main", "TX", "RX":
		return Success, nil
	default:
		return "", fmt.Errorf("no such VFO %s", args[0])
	}
}

func get_split_vfo(ctx context.Context, _ []string) (string, error) {
	return "0\nVFOA\n", nil
}

func set_split_vfo(ctx context.Context, args []string) (string, error) {
	if args[0] == "0" && args[1] == "VFOA" {
		return Success, nil
	} else {
		return "", fmt.Errorf("invalid set split S %s %s", args[0], args[1])
	}
}
