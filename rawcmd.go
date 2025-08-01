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
	res, err := fc.SendAndWaitContext(ctx, cmd)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08X %s\n", res.Error, res.Message), nil
}
