package internal

import (
	"jsonrpc_server/internal/workers"
	"jsonrpc_server/pkg/jsonrpc_server"
)

type Router struct {
}

func (r *Router) ResolveWorker(method string) (jsonrpc_server.Worker, *jsonrpc_server.Error) {
	routes := map[string]jsonrpc_server.Worker{
		"version": &workers.VersionWorker{},
	}

	w, found := routes[method]

	if !found {
		return nil, &jsonrpc_server.Error{Code: -32601, Message: "The method does not exist or it's not available."}
	}

	return w, nil
}
