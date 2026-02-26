package database

import (
	"context"
	"fmt"

	"github.com/atornesi/tsql-ls/dialect"
)

func Open(ctx context.Context, cfg *DBConfig) (DBRepository, error) {
	switch cfg.Driver {
	case dialect.DatabaseDriverMssql:
		return OpenMSSQL(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported driver: %q", cfg.Driver)
	}
}
