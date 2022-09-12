package workers

import "jsonrpc_server/pkg/jsonrpc_server"

type DummyWorker struct {
	Foo int    `json:"foo" validate:"required"`
	Bar string `json:"bar" validate:"required"`
}

func (w *DummyWorker) Handle() (interface{}, *jsonrpc_server.Error) {
	return "success", nil
}
