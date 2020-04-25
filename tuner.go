package main

import "fmt"

func init() {
	hamlib.AddHandler(
		names{{`u`, `TUNER`}, {`\get_func`, `TUNER`}},
		NewHandler(
			get_func_tuner,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`U`, `TUNER`}, {`\set_func`, `TUNER`}},
		NewHandler(
			set_func_tuner,
			Args(1),
		),
	)
}

func get_func_tuner(_ Conn, _ []string) (string, error) {
	xmit, ok := fc.GetObject("transmit")
	if !ok {
		return "", fmt.Errorf("couldn't get transmit object")
	}
	return xmit["tune"] + "\n", nil
}

func set_func_tuner(_ Conn, args []string) (string, error) {
	// TODO: use different values here for TUNE vs ATU
	tune := "1"
	if args[0] == "0" {
		tune = "0"
	}

	res := fc.TransmitTune(tune)
	if res.Error != 0 {
		return "", fmt.Errorf("transmit tune %08X", res.Error)
	}

	return Success, nil
}
