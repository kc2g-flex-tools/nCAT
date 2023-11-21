package main

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	hamlib.AddHandler(
		NewHandler(
			"send_cmd", "w",
			send_cmd,
		),
	)
}

func send_cmd(ctx context.Context, args []string) (string, error) {
	cmd := strings.Join(args, " ")
	res := fc.SendAndWait(cmd)
	return fmt.Sprintf("%08X %s\n", res.Error, res.Message), nil
}
