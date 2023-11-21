package main

import "context"

type connKey struct{}

func getConn(ctx context.Context) *Conn {
	return ctx.Value(connKey{}).(*Conn)
}

type handlerKey struct{}

func getHandler(ctx context.Context) *Handler {
	return ctx.Value(handlerKey{}).(*Handler)
}
