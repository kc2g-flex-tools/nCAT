package main

import "context"

type connKey struct{}

func getConn(ctx context.Context) *Conn {
	return ctx.Value(connKey{}).(*Conn)
}
