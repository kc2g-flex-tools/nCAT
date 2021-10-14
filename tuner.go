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
	atu, ok := fc.GetObject("atu")
	if !ok {
		return "", fmt.Errorf("couldn't get atu object")
	}

	if xmit["tune"] != "0" {
		return "2\n", nil
	} else if atu["status"] == "TUNE_IN_PROGRESS" {
		return "3\n", nil
	} else if atu["status"] != "TUNE_MANUAL_BYPASS" {
		return "1\n", nil
	} else {
		return "0\n", nil
	}
}

func set_func_tuner(_ Conn, args []string) (string, error) {
	disableATU := func() error {
		xmit, ok := fc.GetObject("transmit")
		if !ok {
			return fmt.Errorf("couldn't get transmit object")
		}

		if xmit["tune"] != "0" {
			if res := fc.TransmitTune("0"); res.Error != 0 {
				return fmt.Errorf("transmit tune %08X", res.Error)
			}
		}
		return nil
	}

	stopTune := func() error {
		atu, ok := fc.GetObject("atu")
		if !ok {
			return fmt.Errorf("couldn't get atu object")
		}

		if atu["status"] != "TUNE_MANUAL_BYPASS" {
			if res := fc.SendAndWait("atu bypass"); res.Error != 0 {
				return fmt.Errorf("atu bypass %08X", res.Error)
			}
		}
		return nil
	}

	switch args[0] {
	case "0":
		if err := disableATU(); err != nil {
			return "", err
		}
		if err := stopTune(); err != nil {
			return "", err
		}
		return Success, nil
	case "1":
		if err := stopTune(); err != nil {
			return "", err
		}
		if res := fc.SendAndWait("atu start"); res.Error != 0 {
			return "", fmt.Errorf("atu start %08X", res.Error)
		}
		return Success, nil
	case "2":
		if res := fc.TransmitTune("1"); res.Error != 0 {
			return "", fmt.Errorf("transmit tune %08X", res.Error)
		}
		return Success, nil
	default:
		return "", fmt.Errorf("invalid tune setting")
	}
}
