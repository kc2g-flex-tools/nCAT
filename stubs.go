package main

func noop(_ Conn, _ []string) (string, error) {
	return Success, nil
}

func zero(_ Conn, _ []string) (string, error) {
	return "0\n", nil
}
