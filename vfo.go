package main

import (
	"fmt"
)

func init() {
	hamlib.AddHandler(
		names{{`\chk_vfo`}},
		NewHandler(
			chk_vfo,
			Args(0),
		),
	)
	hamlib.AddHandler(
		names{{`v`}, {`\get_vfo`}},
		NewHandler(
			get_vfo,
			Args(0),
		),
	)
	hamlib.AddHandler(
		names{{`V`}, {`\set_vfo`}},
		NewHandler(
			set_vfo,
			Args(1),
		),
	)
	hamlib.AddHandler(
		names{{`s`}, {`\get_split_vfo`}},
		NewHandler(
			get_split_vfo,
			Args(0),
		),
	)
	hamlib.AddHandler(
		names{{`S`}, {`\set_split_vfo`}},
		NewHandler(
			set_split_vfo,
			Args(2),
		),
	)
}

var swapVFO bool = false

func chk_vfo(_ Conn, _ []string) (string, error) {
	if cfg.ChkVFOMode == "new" {
		return "0\n", nil
	} else {
		return "CHKVFO 0\n", nil
	}
}

func get_vfo(_ Conn, _ []string) (string, error) {
	return "VFOA\n", nil
}

func set_vfo(_ Conn, args []string) (string, error) {
	switch args[0] {
	case "?":
		// List available VFOs
		return "VFOA\n", nil
	case "VFOA", "Main", "RX":
		swapVFO = false
		return Success, nil
	case "VFOB", "Sub":
		if cfg.SplitXIT {
			swapVFO = true
			return Success, nil
		}
	case "TX":
		if cfg.SplitXIT {
			swapVFO = true
		} else {
			swapVFO = false
		}
		return Success, nil
	}
	return "", fmt.Errorf("no such VFO %s", args[0])
}

func get_split_vfo(_ Conn, _ []string) (string, error) {
	if cfg.SplitXIT {
		slice, ok := fc.GetObject("slice " + SliceIdx)
		if !ok {
			return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
		}
		return fmt.Sprintf("%s\nVFOB\n", slice["xit_on"]), nil
	} else {
		return "0\nVFOA\n", nil
	}
}

func set_split_vfo(_ Conn, args []string) (string, error) {
	if args[0] == "0" && args[1] == "VFOA" {
		if cfg.SplitXIT {
			return enable_ritxit("xit", 0)
		} else {
			return Success, nil
		}
	} else if cfg.SplitXIT && args[0] == "1" && (args[1] == "VFOB" || args[1] == "Sub") {
		return enable_ritxit("xit", 1)
	} else {
		return "", fmt.Errorf("invalid set split S %s %s", args[0], args[1])
	}
}
