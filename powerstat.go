package main

import "context"

func init() {
	hamlib.AddHandler(
		NewHandler(
			"get_powerstat", "",
			func(ctx context.Context, _ []string) (string, error) {
				return "1\n", nil
			},
			Args(0),
		),
	)
}
