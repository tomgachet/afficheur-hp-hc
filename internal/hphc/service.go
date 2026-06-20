package hphc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var ErrNoSlot = errors.New("no HP/HC slot found")

type SlotStatus struct {
	Timestamp           time.Time
	CurrentType         string
	CurrentPeriod       string
	CurrentStart        time.Time
	CurrentEnd          time.Time
	RemainingMinutes    int
	NextType            string
	NextPeriod          string
	NextStart           time.Time
	NextEnd             time.Time
	NextDurationMinutes int
}

func CurrentSlot(ctx context.Context, db *sql.DB, at time.Time) (SlotStatus, error) {
	currentSlotSQL, err := readSQLFile("current_slot.sql")
	if err != nil {
		return SlotStatus{}, err
	}

	row := db.QueryRowContext(ctx, currentSlotSQL, at.Format("2006-01-02 15:04:05"))

	var status SlotStatus
	if err := row.Scan(
		&status.Timestamp,
		&status.CurrentType,
		&status.CurrentPeriod,
		&status.CurrentStart,
		&status.CurrentEnd,
		&status.RemainingMinutes,
		&status.NextType,
		&status.NextPeriod,
		&status.NextStart,
		&status.NextEnd,
		&status.NextDurationMinutes,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SlotStatus{}, ErrNoSlot
		}
		return SlotStatus{}, fmt.Errorf("query current slot: %w", err)
	}

	return status, nil
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
