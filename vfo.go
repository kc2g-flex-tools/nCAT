package main

import "fmt"

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
}

func chk_vfo(_ Conn, _ []string) (string, error) {
	return "CHKVFO 0\n", nil
}

func get_vfo(_ Conn, _ []string) (string, error) {
	return "VFOA\n", nil
}

func set_vfo(_ Conn, args []string) (string, error) {
	switch args[0] {
	case "?":
		// List available VFOs
		return "VFOA\n", nil
	case "VFOA":
		return Success, nil
	default:
		return "", fmt.Errorf("no such VFO %s", args[0])
	}
}

func get_split_vfo(_ Conn, _ []string) (string, error) {
	return "0\nVFOA\n", nil
}
