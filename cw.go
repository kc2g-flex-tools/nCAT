package main

import (
	"fmt"
	"strconv"
	"strings"
)

func init() {
	// TODO flex has separate CW WPM and CWX WPM, we should
	// maybe deal with that somehow
	hamlib.AddHandler(
		names{{`l`, `KEYSPD`}, {`\get_level`, `KEYSPD`}},
		NewHandler(
			get_level_keyspd,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`L`, `KEYSPD`}, {`\set_level`, `KEYSPD`}},
		NewHandler(
			set_level_keyspd,
			Args(1),
		),
	)
	hamlib.AddHandler(
		names{{`l`, `CWPITCH`}, {`\get_level`, `CWPITCH`}},
		NewHandler(
			get_level_cwpitch,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`L`, `CWPITCH`}, {`\set_level`, `CWPITCH`}},
		NewHandler(
			set_level_cwpitch,
			Args(1),
		),
	)

	hamlib.AddHandler(
		names{{`l`, `BKINDL`}, {`\get_level`, `BKINDL`}},
		NewHandler(
			get_level_bkindl,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`L`, `BKINDL`}, {`\set_level`, `BKINDL`}},
		NewHandler(
			set_level_cwpitch,
			Args(1),
		),
	)

	hamlib.AddHandler(
		names{{`u`, `SBKIN`}, {`\get_func`, `SBKIN`},
			{`u`, `FBKIN`}, {`\get_func`, `FBKIN`}},
		NewHandler(
			get_func_bkin,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`U`, `SBKIN`}, {`\set_func`, `SBKIN`},
			{`U`, `FBKIN`}, {`\set_func`, `FBKIN`}},
		NewHandler(
			set_func_bkin,
			Args(0),
		),
	)

	hamlib.AddHandler(
		names{{`b`}, {`\send_morse`}},
		NewHandler(
			send_morse,
		),
	)

}

func get_level_keyspd(_ Conn, _ []string) (string, error) {
	xmit, ok := fc.GetObject("transmit")
	if !ok {
		return "", fmt.Errorf("couldn't get transmit obj")
	}
	return xmit["wpm"] + "\n", nil
}

func set_level_keyspd(_ Conn, args []string) (string, error) {
	keyspd, err := strconv.Atoi(args[0])
	if err != nil {
		return "", err
	}
	res := fc.SendAndWait(fmt.Sprintf("cw wpm %d", keyspd))
	if res.Error != 0 {
		return "", fmt.Errorf("cw wpm %08X", res.Error)
	}
	return Success, nil
}

func get_level_cwpitch(_ Conn, _ []string) (string, error) {
	xmit, ok := fc.GetObject("transmit")
	if !ok {
		return "", fmt.Errorf("couldn't get transmit obj")
	}
	return xmit["pitch"] + "\n", nil
}

func set_level_cwpitch(_ Conn, args []string) (string, error) {
	cwpitch, err := strconv.Atoi(args[0])
	if err != nil {
		return "", err
	}

	if cwpitch < 100 {
		cwpitch = 100
	}
	if cwpitch > 6000 {
		cwpitch = 6000
	}

	res := fc.SendAndWait(fmt.Sprintf("cw pitch %d", cwpitch))
	if res.Error != 0 {
		return "", fmt.Errorf("cw pitch %08X", res.Error)
	}
	return Success, nil
}

func get_level_bkindl(_ Conn, _ []string) (string, error) {
	xmit, ok := fc.GetObject("transmit")
	if !ok {
		return "", fmt.Errorf("couldn't get transmit obj")
	}
	wpm, err := strconv.Atoi(xmit["wpm"])
	if err != nil {
		return "", err
	}
	bkindl, err := strconv.Atoi(xmit["break_in_delay"])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.0f\n", 120*float64(bkindl)/float64(wpm)), nil
}

func set_level_bkindl(_ Conn, args []string) (string, error) {
	bkindl, err := strconv.Atoi(args[0])
	if err != nil {
		return "", err
	}
	xmit, ok := fc.GetObject("transmit")
	if !ok {
		return "", fmt.Errorf("couldn't get transmit obj")
	}
	wpm, err := strconv.Atoi(xmit["wpm"])
	if err != nil {
		return "", err
	}

	ms := float64(bkindl) * float64(wpm) / 120
	if ms > 2000 {
		ms = 2000
	}

	res := fc.SendAndWait(fmt.Sprintf("cw break_in_delay %.0f", ms))
	if res.Error != 0 {
		return "", fmt.Errorf("cw break_in_delay %08X", res.Error)
	}
	return Success, nil
}

func get_func_bkin(_ Conn, _ []string) (string, error) {
	xmit, ok := fc.GetObject("transmit")
	if !ok {
		return "", fmt.Errorf("couldn't get transmit obj")
	}
	return xmit["break_in"] + "\n", nil
}

func set_func_bkin(_ Conn, args []string) (string, error) {
	if args[0] != "0" && args[0] != "1" {
		return "", fmt.Errorf("invalid breakin value")
	}

	res := fc.SendAndWait(fmt.Sprintf("cw break_in %s", args[0]))

	if res.Error != 0 {
		return "", fmt.Errorf("cw break_in %08X", res.Error)
	}
	return Success, nil
}

func send_morse(_ Conn, args []string) (string, error) {
	text := strings.Join(args, " ")
	text = strings.ReplaceAll(text, " ", "\u007f")

	res := fc.SendAndWait(fmt.Sprintf("cwx send \"%s\"", text))
	if res.Error != 0 {
		return "", fmt.Errorf("cwx send %08X", res.Error)
	}
	return Success, nil
}
