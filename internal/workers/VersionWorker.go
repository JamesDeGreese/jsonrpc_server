package workers

import (
	"context"
	"jsonrpc_server/internal/core"
	"jsonrpc_server/pkg/jsonrpc_server"
)

type VersionWorker struct {
	Foo int    `json:"foo" validate:"required"`
	Bar string `json:"bar" validate:"required"`
}

func (w *VersionWorker) Handle(ctx context.Context) (interface{}, *jsonrpc_server.Error) {
	app := ctx.Value("app").(*core.App)
	return app.Config.AppVersion, nil
}
