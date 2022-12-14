package main

import (
	"context"
	"jsonrpc_server/internal"
	"jsonrpc_server/internal/core"
	"jsonrpc_server/pkg/jsonrpc_server"
)

func main() {
	app, err := core.Init("config/app.toml")

	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	ctx = context.WithValue(ctx, "app", app)

	s := jsonrpc_server.Server{
		Address: app.Config.Http.Address,
		Router:  &internal.Router{},
		Ctx:     ctx,
		Timeout: app.Config.Http.Timeout,
	}

	s.Run()
}
