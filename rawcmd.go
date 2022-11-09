package main

import (
	"fmt"
	"strings"
)

func init() {
	hamlib.AddHandler(
		names{{`w`}, {`\send_cmd`}},
		NewHandler(
			send_cmd,
		),
	)
}

func send_cmd(_ *Conn, args []string) (string, error) {
	cmd := strings.Join(args, " ")
	res := fc.SendAndWait(cmd)
	return fmt.Sprintf("%08X %s\n", res.Error, res.Message), nil
}
