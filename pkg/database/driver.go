package database

import (
	"github.com/atornesi/tsql-ls/dialect"
)

type DBConnection struct {
	Driver dialect.DatabaseDriver
}

func (db *DBConnection) Close() error {
	return nil
}
