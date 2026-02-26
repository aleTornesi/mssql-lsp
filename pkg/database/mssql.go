package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"

	_ "github.com/microsoft/go-mssqldb"

	"github.com/aleTornesi/mssql-lsp/dialect"
)

type MSSQLDBRepository struct {
	db *sql.DB
}

func OpenMSSQL(ctx context.Context, cfg *DBConfig) (*MSSQLDBRepository, error) {
	dsn, err := mssqlDSN(cfg)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return nil, fmt.Errorf("mssql open: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("mssql ping: %w", err)
	}

	db.SetMaxIdleConns(DefaultMaxIdleConns)
	db.SetMaxOpenConns(DefaultMaxOpenConns)

	return &MSSQLDBRepository{db: db}, nil
}

func mssqlDSN(cfg *DBConfig) (string, error) {
	if cfg.DataSourceName != "" {
		return cfg.DataSourceName, nil
	}

	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Port
	if port == 0 {
		port = 1433
	}

	q := url.Values{}
	if cfg.DBName != "" {
		q.Set("database", cfg.DBName)
	}
	for k, v := range cfg.Params {
		q.Set(k, v)
	}

	var userInfo *url.Userinfo
	if cfg.User != "" {
		userInfo = url.UserPassword(cfg.User, cfg.Passwd)
	} else {
		q.Set("integrated security", "true")
	}

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     userInfo,
		Host:     host + ":" + strconv.Itoa(port),
		RawQuery: q.Encode(),
	}
	return u.String(), nil
}

func (r *MSSQLDBRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *MSSQLDBRepository) Driver() dialect.DatabaseDriver {
	return dialect.DatabaseDriverMssql
}

func (r *MSSQLDBRepository) CurrentDatabase(ctx context.Context) (string, error) {
	var name string
	if err := r.db.QueryRowContext(ctx, "SELECT DB_NAME()").Scan(&name); err != nil {
		return "", err
	}
	return name, nil
}

func (r *MSSQLDBRepository) CurrentSchema(ctx context.Context) (string, error) {
	var name string
	if err := r.db.QueryRowContext(ctx, "SELECT SCHEMA_NAME()").Scan(&name); err != nil {
		return "", err
	}
	return name, nil
}

func (r *MSSQLDBRepository) Databases(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT name FROM sys.databases ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbs []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		dbs = append(dbs, name)
	}
	return dbs, rows.Err()
}

func (r *MSSQLDBRepository) Schemas(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.name
		FROM sys.schemas s
		JOIN sys.database_principals p ON s.principal_id = p.principal_id
		WHERE p.type IN ('S','U','G')
		ORDER BY s.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		schemas = append(schemas, name)
	}
	return schemas, rows.Err()
}

func (r *MSSQLDBRepository) SchemaTables(ctx context.Context) (map[string][]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT TABLE_SCHEMA, TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		ORDER BY TABLE_SCHEMA, TABLE_NAME`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var schema, table string
		if err := rows.Scan(&schema, &table); err != nil {
			return nil, err
		}
		result[schema] = append(result[schema], table)
	}
	return result, rows.Err()
}

func (r *MSSQLDBRepository) DescribeDatabaseTable(ctx context.Context) ([]*ColumnDesc, error) {
	return r.describeColumns(ctx, "")
}

func (r *MSSQLDBRepository) DescribeDatabaseTableBySchema(ctx context.Context, schemaName string) ([]*ColumnDesc, error) {
	return r.describeColumns(ctx, schemaName)
}

func (r *MSSQLDBRepository) describeColumns(ctx context.Context, schemaName string) ([]*ColumnDesc, error) {
	query := `
		SELECT
			c.TABLE_SCHEMA,
			c.TABLE_NAME,
			c.COLUMN_NAME,
			c.DATA_TYPE
				+ CASE
					WHEN c.CHARACTER_MAXIMUM_LENGTH IS NOT NULL
						THEN '(' + CAST(c.CHARACTER_MAXIMUM_LENGTH AS VARCHAR) + ')'
					WHEN c.NUMERIC_PRECISION IS NOT NULL AND c.DATA_TYPE NOT IN ('int','bigint','smallint','tinyint','float','real','money','smallmoney','bit')
						THEN '(' + CAST(c.NUMERIC_PRECISION AS VARCHAR) + ',' + CAST(ISNULL(c.NUMERIC_SCALE,0) AS VARCHAR) + ')'
					ELSE ''
				END AS DATA_TYPE,
			c.IS_NULLABLE,
			CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 'YES' ELSE 'NO' END AS IS_PRIMARY_KEY,
			c.COLUMN_DEFAULT,
			CASE
				WHEN sc.is_identity = 1 THEN 'IDENTITY'
				WHEN cc.definition IS NOT NULL THEN 'COMPUTED: ' + cc.definition
				ELSE ''
			END AS EXTRA
		FROM INFORMATION_SCHEMA.COLUMNS c
		LEFT JOIN (
			SELECT ku.TABLE_SCHEMA, ku.TABLE_NAME, ku.COLUMN_NAME
			FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku
				ON tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
				AND tc.TABLE_SCHEMA = ku.TABLE_SCHEMA
			WHERE tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
		) pk ON c.TABLE_SCHEMA = pk.TABLE_SCHEMA
			AND c.TABLE_NAME = pk.TABLE_NAME
			AND c.COLUMN_NAME = pk.COLUMN_NAME
		LEFT JOIN sys.columns sc
			ON sc.object_id = OBJECT_ID(QUOTENAME(c.TABLE_SCHEMA) + '.' + QUOTENAME(c.TABLE_NAME))
			AND sc.name = c.COLUMN_NAME
		LEFT JOIN sys.computed_columns cc
			ON cc.object_id = sc.object_id
			AND cc.column_id = sc.column_id
		WHERE 1=1`

	var args []interface{}
	if schemaName != "" {
		query += " AND c.TABLE_SCHEMA = @p1"
		args = append(args, schemaName)
	}
	query += " ORDER BY c.TABLE_SCHEMA, c.TABLE_NAME, c.ORDINAL_POSITION"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []*ColumnDesc
	for rows.Next() {
		col := &ColumnDesc{}
		if err := rows.Scan(
			&col.Schema,
			&col.Table,
			&col.Name,
			&col.Type,
			&col.Null,
			&col.Key,
			&col.Default,
			&col.Extra,
		); err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	return cols, rows.Err()
}

func (r *MSSQLDBRepository) DescribeForeignKeysBySchema(ctx context.Context, schemaName string) ([]*ForeignKey, error) {
	query := `
		SELECT
			fk.name AS fk_name,
			tp.name AS parent_table,
			cp.name AS parent_column,
			tr.name AS referenced_table,
			cr.name AS referenced_column
		FROM sys.foreign_keys fk
		JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
		JOIN sys.tables tp ON fkc.parent_object_id = tp.object_id
		JOIN sys.columns cp ON fkc.parent_object_id = cp.object_id AND fkc.parent_column_id = cp.column_id
		JOIN sys.tables tr ON fkc.referenced_object_id = tr.object_id
		JOIN sys.columns cr ON fkc.referenced_object_id = cr.object_id AND fkc.referenced_column_id = cr.column_id
		JOIN sys.schemas s ON tp.schema_id = s.schema_id
		WHERE s.name = @p1
		ORDER BY fk.name, fkc.constraint_column_id`

	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return parseForeignKeys(rows, schemaName)
}

func (r *MSSQLDBRepository) Exec(ctx context.Context, query string) (sql.Result, error) {
	return r.db.ExecContext(ctx, query)
}

func (r *MSSQLDBRepository) Query(ctx context.Context, query string) (*sql.Rows, error) {
	return r.db.QueryContext(ctx, query)
}
