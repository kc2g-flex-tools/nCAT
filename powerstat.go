package main

import "context"

func init() {
	hamlib.AddHandler(
		names{{`\get_powerstat`}},
		NewHandler(
			func(ctx context.Context, _ []string) (string, error) {
				return "1\n", nil
			},
			Args(0),
		),
	)
}
