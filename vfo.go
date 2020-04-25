package main

import "fmt"

func init() {
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
}

func get_vfo(_ Conn, _ []string) (string, error) {
	return "VFOA\n", nil
}

func set_vfo(_ Conn, args []string) (string, error) {
	if args[0] == "?" {
		return "VFOA\n", nil
	} else if args[0] == "VFOA" {
		return Success, nil
	} else {
		return "", fmt.Errorf("No such VFO %s", args[0])
	}
}
