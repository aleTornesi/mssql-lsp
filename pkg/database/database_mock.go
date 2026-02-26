package database

import (
	"context"
	"database/sql"

	"github.com/aleTornesi/mssql-lsp/dialect"
)

type MockDBRepository struct {
	MockDatabase                      func(context.Context) (string, error)
	MockDatabases                     func(context.Context) ([]string, error)
	MockDatabaseTables                func(context.Context) (map[string][]string, error)
	MockDescribeDatabaseTable         func(context.Context) ([]*ColumnDesc, error)
	MockDescribeDatabaseTableBySchema func(context.Context, string) ([]*ColumnDesc, error)
	MockExec                          func(context.Context, string) (sql.Result, error)
	MockQuery                         func(context.Context, string) (*sql.Rows, error)
	MockDescribeForeignKeysBySchema   func(context.Context, string) ([]*ForeignKey, error)
}

func NewMockDBRepository() DBRepository {
	return &MockDBRepository{
		MockDatabase:       func(ctx context.Context) (string, error) { return "dbo", nil },
		MockDatabases:      func(ctx context.Context) ([]string, error) { return dummyDatabases, nil },
		MockDatabaseTables: func(ctx context.Context) (map[string][]string, error) { return dummyDatabaseTables, nil },
		MockDescribeDatabaseTable: func(ctx context.Context) ([]*ColumnDesc, error) {
			var res []*ColumnDesc
			res = append(res, dummyEmployeeColumns...)
			res = append(res, dummyDepartmentColumns...)
			return res, nil
		},
		MockDescribeDatabaseTableBySchema: func(ctx context.Context, schemaName string) ([]*ColumnDesc, error) {
			var res []*ColumnDesc
			res = append(res, dummyEmployeeColumns...)
			res = append(res, dummyDepartmentColumns...)
			return res, nil
		},
		MockExec: func(ctx context.Context, query string) (sql.Result, error) {
			return nil, nil
		},
		MockQuery: func(ctx context.Context, query string) (*sql.Rows, error) {
			return nil, nil
		},
		MockDescribeForeignKeysBySchema: func(ctx context.Context, schemaName string) ([]*ForeignKey, error) {
			return foreignKeys, nil
		},
	}
}

func (m *MockDBRepository) Driver() dialect.DatabaseDriver {
	return "mock"
}

func (m *MockDBRepository) CurrentDatabase(ctx context.Context) (string, error) {
	return m.MockDatabase(ctx)
}

func (m *MockDBRepository) Databases(ctx context.Context) ([]string, error) {
	return m.MockDatabases(ctx)
}

func (m *MockDBRepository) CurrentSchema(ctx context.Context) (string, error) {
	return m.MockDatabase(ctx)
}

func (m *MockDBRepository) Schemas(ctx context.Context) ([]string, error) {
	return m.MockDatabases(ctx)
}

func (m *MockDBRepository) SchemaTables(ctx context.Context) (map[string][]string, error) {
	return m.MockDatabaseTables(ctx)
}

func (m *MockDBRepository) DescribeDatabaseTable(ctx context.Context) ([]*ColumnDesc, error) {
	return m.MockDescribeDatabaseTable(ctx)
}

func (m *MockDBRepository) DescribeDatabaseTableBySchema(ctx context.Context, schemaName string) ([]*ColumnDesc, error) {
	return m.MockDescribeDatabaseTableBySchema(ctx, schemaName)
}

func (m *MockDBRepository) Exec(ctx context.Context, query string) (sql.Result, error) {
	return m.MockExec(ctx, query)
}

func (m *MockDBRepository) Query(ctx context.Context, query string) (*sql.Rows, error) {
	return m.MockQuery(ctx, query)
}

func (m *MockDBRepository) DescribeForeignKeysBySchema(ctx context.Context, schemaName string) ([]*ForeignKey, error) {
	return m.MockDescribeForeignKeysBySchema(ctx, schemaName)
}

var dummyDatabases = []string{
	"dbo",
}

var dummyDatabaseTables = map[string][]string{
	"dbo": {
		"Employee",
		"Department",
	},
}

var dummyEmployeeColumns = []*ColumnDesc{
	{
		ColumnBase: ColumnBase{Schema: "dbo", Table: "Employee", Name: "EmployeeID"},
		Type:       "int",
		Null:       "NO",
		Key:        "YES",
		Default:    sql.NullString{Valid: false},
		Extra:      "IDENTITY",
	},
	{
		ColumnBase: ColumnBase{Schema: "dbo", Table: "Employee", Name: "FirstName"},
		Type:       "nvarchar(50)",
		Null:       "NO",
		Key:        "",
		Default:    sql.NullString{Valid: false},
		Extra:      "",
	},
	{
		ColumnBase: ColumnBase{Schema: "dbo", Table: "Employee", Name: "LastName"},
		Type:       "nvarchar(50)",
		Null:       "NO",
		Key:        "",
		Default:    sql.NullString{Valid: false},
		Extra:      "",
	},
	{
		ColumnBase: ColumnBase{Schema: "dbo", Table: "Employee", Name: "DepartmentID"},
		Type:       "int",
		Null:       "YES",
		Key:        "",
		Default:    sql.NullString{Valid: false},
		Extra:      "",
	},
	{
		ColumnBase: ColumnBase{Schema: "dbo", Table: "Employee", Name: "Salary"},
		Type:       "decimal(10,2)",
		Null:       "YES",
		Key:        "",
		Default:    sql.NullString{Valid: false},
		Extra:      "",
	},
}

var dummyDepartmentColumns = []*ColumnDesc{
	{
		ColumnBase: ColumnBase{Schema: "dbo", Table: "Department", Name: "DepartmentID"},
		Type:       "int",
		Null:       "NO",
		Key:        "YES",
		Default:    sql.NullString{Valid: false},
		Extra:      "IDENTITY",
	},
	{
		ColumnBase: ColumnBase{Schema: "dbo", Table: "Department", Name: "DepartmentName"},
		Type:       "nvarchar(100)",
		Null:       "NO",
		Key:        "",
		Default:    sql.NullString{Valid: false},
		Extra:      "",
	},
}

var foreignKeys = []*ForeignKey{
	{
		[2]*ColumnBase{
			{Schema: "dbo", Table: "Employee", Name: "DepartmentID"},
			{Schema: "dbo", Table: "Department", Name: "DepartmentID"},
		},
	},
}
