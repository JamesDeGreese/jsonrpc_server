package main

import (
	"jsonrpc_server/internal"
	"jsonrpc_server/pkg/jsonrpc_server"
)

func main() {
	s := jsonrpc_server.Server{
		Address: ":80",
		Router:  &internal.Router{},
	}

	s.Run()
}
