package core

import (
	"context"
	"github.com/jackc/pgx/v4"
	"jsonrpc_server/internal/core/config"
)

type App struct {
	Database *pgx.Conn
	Config   config.Config
}

func Init(configFile string) (*App, error) {
	app := App{}

	cfg, err := config.Parse(configFile)

	if err != nil {
		return &app, err
	}

	app.Config = cfg

	DBConn, err := pgx.Connect(context.Background(), cfg.Database.GetDatabaseURL())
	if err != nil {
		return &app, err
	}

	app.Database = DBConn

	defer DBConn.Close(context.Background())

	return &app, nil
}
