package main

import (
	"fmt"
	"strconv"

	"github.com/kc2g-flex-tools/flexclient"
)

func init() {
	hamlib.AddHandler(
		names{{`j`}, {`\get_rit`}},
		NewHandler(
			get_rit,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`J`}, {`\set_rit`}},
		NewHandler(
			set_rit,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`u`, `RIT`}, {`\get_func`, `RIT`}},
		NewHandler(
			get_func_rit,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`U`, `RIT`}, {`\set_func`, `RIT`}},
		NewHandler(
			set_func_rit,
			Args(1),
		),
	)

	hamlib.AddHandler(
		names{{`z`}, {`\get_xit`}},
		NewHandler(
			get_xit,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`Z`}, {`\set_xit`}},
		NewHandler(
			set_xit,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`u`, `XIT`}, {`\get_func`, `XIT`}},
		NewHandler(
			get_func_xit,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`U`, `XIT`}, {`\set_func`, `XIT`}},
		NewHandler(
			set_func_xit,
			Args(1),
		),
	)
}

func enable_ritxit(ctl string, enabled int) (string, error) {
	obj := flexclient.Object{}
	obj[ctl+"_on"] = fmt.Sprintf("%d", enabled)
	res := fc.SliceSet(SliceIdx, obj)
	if res.Error != 0 {
		return "", fmt.Errorf("slice set %08X", res.Error)
	}
	return Success, nil
}

func set_ritxit_freq(ctl string, offset int64) (string, error) {
	if offset < -99999 || offset > 99999 {
		return "", fmt.Errorf("rit offset %d out of range", offset)
	}
	obj := flexclient.Object{}
	obj[ctl+"_freq"] = fmt.Sprintf("%d", offset)
	res := fc.SliceSet(SliceIdx, obj)
	if res.Error != 0 {
		return "", fmt.Errorf("slice set %08X", res.Error)
	}
	return Success, nil
}

func get_func_rit(_ Conn, _ []string) (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}
	return slice["rit"] + "\n", nil
}

func set_func_rit(_ Conn, args []string) (string, error) {
	enabled, err := strconv.Atoi(args[0])
	if err != nil {
		return "", err
	}
	return enable_ritxit("rit", enabled)
}

func get_rit(_ Conn, _ []string) (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}
	return slice["rit_freq"] + "\n", nil
}

func set_rit(_ Conn, args []string) (string, error) {
	offset, err := strconv.Atoi(args[0])
	if err != nil {
		return "", err
	}
	return set_ritxit_freq("rit", int64(offset))
}

func get_func_xit(_ Conn, _ []string) (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}
	return slice["xit"] + "\n", nil
}

func set_func_xit(_ Conn, args []string) (string, error) {
	enabled, err := strconv.Atoi(args[0])
	if err != nil {
		return "", err
	}
	return enable_ritxit("xit", enabled)
}

func get_xit(_ Conn, _ []string) (string, error) {
	slice, ok := fc.GetObject("slice " + SliceIdx)
	if !ok {
		return "", fmt.Errorf("couldn't get slice %s", SliceIdx)
	}
	return slice["xit_freq"] + "\n", nil
}

func set_xit(_ Conn, args []string) (string, error) {
	offset, err := strconv.Atoi(args[0])
	if err != nil {
		return "", err
	}
	return set_ritxit_freq("xit", int64(offset))
}
