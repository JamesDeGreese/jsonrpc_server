package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

type Database struct {
	Driver   string `toml:"driver"`
	UserName string `toml:"user"`
	Password string `toml:"password"`
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	Database string `toml:"dbName"`
}

type Http struct {
	Address string `toml:"address"`
}

type Config struct {
	Database   Database `toml:"database"`
	Http       Http     `toml:"http"`
	AppVersion string   `toml:"version"`
}

func Parse(fileName string) (Config, error) {
	var cfg Config

	_, err := toml.DecodeFile(fileName, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (db *Database) GetDatabaseURL() string {
	return fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=disable",
		db.Driver,
		db.UserName,
		db.Password,
		db.Host,
		db.Port,
		db.Database,
	)
}
