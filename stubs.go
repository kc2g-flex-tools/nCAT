package main

import "context"

func noop(ctx context.Context, _ []string) (string, error) {
	return Success, nil
}

func zero(ctx context.Context, _ []string) (string, error) {
	return "0\n", nil
}
