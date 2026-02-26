package database

import (
	"github.com/aleTornesi/mssql-lsp/dialect"
)

type DBConnection struct {
	Driver dialect.DatabaseDriver
}

func (db *DBConnection) Close() error {
	return nil
}
