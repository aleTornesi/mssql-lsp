package database

import (
	"github.com/atornesi/tsql-ls/dialect"
)

type Proto string

const (
	ProtoTCP Proto = "tcp"
)

type DBConfig struct {
	Alias          string                 `json:"alias" yaml:"alias"`
	Driver         dialect.DatabaseDriver `json:"driver" yaml:"driver"`
	DataSourceName string                 `json:"dataSourceName" yaml:"dataSourceName"`
	Proto          Proto                  `json:"proto" yaml:"proto"`
	User           string                 `json:"user" yaml:"user"`
	Passwd         string                 `json:"passwd" yaml:"passwd"`
	Host           string                 `json:"host" yaml:"host"`
	Port           int                    `json:"port" yaml:"port"`
	DBName         string                 `json:"dbName" yaml:"dbName"`
	Params         map[string]string      `json:"params" yaml:"params"`
}
