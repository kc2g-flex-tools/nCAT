package main

import "fmt"

func init() {
	hamlib.AddHandler(
		names{{`t`}, {`\get_ptt`}},
		NewHandler(
			get_ptt,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`T`}, {`\set_ptt`}},
		NewHandler(
			set_ptt,
			Args(1),
		),
	)
}

func get_ptt(_ Conn, _ []string) (string, error) {
	interlock, ok := fc.GetObject("interlock")
	if !ok {
		return "", fmt.Errorf("couldn't get interlock")
	}

	if interlock["state"] == "TRANSMITTING" {
		return "1\n", nil
	} else {
		return "0\n", nil
	}
}

func set_ptt(_ Conn, args []string) (string, error) {
	tx := "1"
	if args[0] == "0" {
		tx = "0"
	}
	res := fc.SendAndWait("xmit " + tx)
	if res.Error != 0 {
		return "", fmt.Errorf("xmit %08X", res.Error)
	}
	return Success, nil
}
