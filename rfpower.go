package main

import (
	"fmt"
	"strconv"

	"github.com/kc2g-flex-tools/flexclient"
)

func init() {
	hamlib.AddHandler(
		names{{`l`, `RFPOWER`}, {`\get_level`, `RFPOWER`}},
		NewHandler(
			get_level_rfpower,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`L`, `RFPOWER`}, {`\set_level`, `RFPOWER`}},
		NewHandler(
			set_level_rfpower,
			Args(1),
		),
	)

	hamlib.AddHandler(
		names{{`l`, `RF`}, {`\get_level`, `RF`}},
		NewHandler(
			get_level_rf,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`L`, `RF`}, {`\set_level`, `RF`}},
		NewHandler(
			set_level_rf,
			Args(1),
		),
	)
}

func get_level_rfpower(_ Conn, _ []string) (string, error) {
	xmit, ok := fc.GetObject("transmit")
	if !ok {
		return "", fmt.Errorf("couldn't get transmit obj")
	}
	power, err := strconv.ParseFloat(xmit["rfpower"], 64)
	if err != nil {
		return "", err
	}
	power /= 100
	return fmt.Sprintf("%.3f\n", power), nil
}

func set_level_rfpower(_ Conn, args []string) (string, error) {
	power, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return "", err
	}
	power *= 100
	res := fc.TransmitSet(flexclient.Object{"rfpower": fmt.Sprintf("%.0f", power)})
	if res.Error != 0 {
		return "", fmt.Errorf("transmit set %08X", res.Error)
	}
	return Success, nil
}

func get_level_rf(_ Conn, _ []string) (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}

	agcKey := "agc_threshold"
	if slice["agc_mode"] == "off" {
		agcKey = "agc_off_level"
	}

	agct, err := strconv.ParseFloat(slice[agcKey], 64)
	if err != nil {
		return "", err
	}
	agct /= 100
	return fmt.Sprintf("%.3f\n", agct), nil
}

func set_level_rf(_ Conn, args []string) (string, error) {
	agct, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return "", err
	}
	agct *= 100

	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}

	obj := flexclient.Object{}
	val := fmt.Sprintf("%.0f", agct)
	if slice["agc_mode"] == "off" {
		obj["agc_off_level"] = val
	} else {
		obj["agc_threshold"] = val
	}

	res := fc.SliceSet(SliceIdx, obj)
	if res.Error != 0 {
		return "", fmt.Errorf("slice set %08X", res.Error)
	}

	return Success, nil
}
