package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/marcboeker/go-duckdb"
)

func Open(ctx context.Context, path string, reloadReference bool) (*sql.DB, error) {
	if path == "" {
		path = ":memory:"
	}

	conn, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}

	if err := Init(ctx, conn, reloadReference); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func Init(ctx context.Context, conn *sql.DB, reloadReference bool) error {
	refTableExists, err := tableExists(ctx, conn, "reference", "ref_time_slot")
	if err != nil {
		return err
	}

	schemaSQL, err := readSQLFile("schema.sql")
	if err != nil {
		return err
	}

	if _, err := conn.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}

	if !refTableExists || reloadReference {
		if err := loadReferenceCSV(ctx, conn); err != nil {
			return err
		}
	}

	return nil
}

func tableExists(ctx context.Context, conn *sql.DB, schemaName, tableName string) (bool, error) {
	var exists bool
	err := conn.QueryRowContext(ctx, `
SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = ?
      AND table_name = ?
)`, schemaName, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check table %s.%s: %w", schemaName, tableName, err)
	}
	return exists, nil
}

func readSQLFile(name string) (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("locate sql file %s: runtime caller unavailable", name)
	}

	path := filepath.Join(filepath.Dir(file), "..", "..", "sql", name)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read sql file %s: %w", name, err)
	}
	return string(content), nil
}

func loadReferenceCSV(ctx context.Context, conn *sql.DB) error {
	csvPath, err := projectFile("ressources", "ref_time_slot.csv")
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
TRUNCATE TABLE reference.ref_time_slot;

INSERT INTO reference.ref_time_slot (
    day_of_week,
    month_of_year,
    start_time,
    end_time,
    period_type,
    pricing_type
)
SELECT
    CAST(day_of_week AS SMALLINT),
    CAST(month_of_year AS SMALLINT),
    CAST(start_time AS TIME),
    CAST(end_time AS TIME),
    period_type,
    pricing_type
FROM read_csv_auto(%s, header = true);
`, sqlString(csvPath))

	if _, err := conn.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("load reference csv %s: %w", csvPath, err)
	}
	return nil
}

func projectFile(parts ...string) (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("locate project file %v: runtime caller unavailable", parts)
	}
	return filepath.Abs(filepath.Join(append([]string{filepath.Dir(file), "..", ".."}, parts...)...))
}

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
