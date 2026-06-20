package db_test

import (
	"context"
	"database/sql"
	"testing"

	appdb "afficheur-hp-hc/internal/db"
)

func TestOpenLoadsReferenceOnlyWhenCreatedOrReloaded(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/test.duckdb"

	conn, err := appdb.Open(ctx, dbPath, false)
	if err != nil {
		t.Fatal(err)
	}

	if got := countRows(t, ctx, conn); got == 0 {
		t.Fatal("expected initial reference load")
	}

	if _, err := conn.ExecContext(ctx, "DELETE FROM reference.ref_time_slot"); err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}

	conn, err = appdb.Open(ctx, dbPath, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := countRows(t, ctx, conn); got != 0 {
		t.Fatalf("rows after open without reload = %d, want 0", got)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}

	conn, err = appdb.Open(ctx, dbPath, true)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if got := countRows(t, ctx, conn); got == 0 {
		t.Fatal("expected forced reference reload")
	}
}

func countRows(t *testing.T, ctx context.Context, conn *sql.DB) int {
	t.Helper()

	var count int
	if err := conn.QueryRowContext(ctx, "SELECT count(*) FROM reference.ref_time_slot").Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
