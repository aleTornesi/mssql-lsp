package database

import (
	"context"
	"testing"
	"time"

	"github.com/atornesi/tsql-ls/dialect"
)

func TestMssqlDSN_Passthrough(t *testing.T) {
	cfg := &DBConfig{
		DataSourceName: "sqlserver://sa:pass@myhost:1433?database=mydb",
	}
	got, err := mssqlDSN(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got != cfg.DataSourceName {
		t.Errorf("expected passthrough DSN %q, got %q", cfg.DataSourceName, got)
	}
}

func TestMssqlDSN_Fields(t *testing.T) {
	cfg := &DBConfig{
		User:   "sa",
		Passwd: "Secret123",
		Host:   "dbserver",
		Port:   1434,
		DBName: "testdb",
	}
	got, err := mssqlDSN(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// Should contain user, host:port, database param
	for _, want := range []string{"sqlserver://", "sa:", "dbserver:1434", "database=testdb"} {
		if !contains(got, want) {
			t.Errorf("DSN %q missing %q", got, want)
		}
	}
}

func TestMssqlDSN_IntegratedAuth(t *testing.T) {
	cfg := &DBConfig{
		Host:   "dbserver",
		DBName: "testdb",
	}
	got, err := mssqlDSN(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(got, "integrated+security=true") {
		t.Errorf("DSN %q missing integrated security", got)
	}
	// Should not contain user info segment
	if contains(got, "sa:") {
		t.Errorf("DSN %q should not have user credentials", got)
	}
}

func TestMssqlDSN_Defaults(t *testing.T) {
	cfg := &DBConfig{}
	got, err := mssqlDSN(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(got, "localhost:1433") {
		t.Errorf("DSN %q missing default host:port", got)
	}
}

func TestMssqlDSN_CustomParams(t *testing.T) {
	cfg := &DBConfig{
		User:   "sa",
		Passwd: "pass",
		Host:   "host",
		Params: map[string]string{
			"encrypt":              "true",
			"TrustServerCertificate": "true",
		},
	}
	got, err := mssqlDSN(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(got, "encrypt=true") {
		t.Errorf("DSN %q missing encrypt param", got)
	}
	if !contains(got, "TrustServerCertificate=true") {
		t.Errorf("DSN %q missing TrustServerCertificate param", got)
	}
}

func TestOpen_UnsupportedDriver(t *testing.T) {
	cfg := &DBConfig{Driver: "postgres"}
	_, err := Open(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
	if !contains(err.Error(), "unsupported driver") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpen_MssqlInvalidHost(t *testing.T) {
	cfg := &DBConfig{
		Driver: dialect.DatabaseDriverMssql,
		Host:   "invalid-host-that-does-not-exist",
		Port:   1433,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := Open(ctx, cfg)
	if err == nil {
		t.Fatal("expected error connecting to invalid host")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
