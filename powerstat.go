package main

func init() {
	hamlib.AddHandler(
		names{{`\get_powerstat`}},
		NewHandler(
			func(_ Conn, _ []string) (string, error) {
				return "1\n", nil
			},
			Args(0),
		),
	)
}
