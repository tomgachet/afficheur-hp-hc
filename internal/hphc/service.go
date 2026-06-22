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

type Slot struct {
	Type            string
	Period          string
	Start           time.Time
	End             time.Time
	DurationMinutes int
}

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

func UpcomingSlots(ctx context.Context, db *sql.DB, at time.Time, limit int) ([]Slot, error) {
	if limit <= 0 {
		return nil, nil
	}

	upcomingSlotsSQL, err := readSQLFile("upcoming_slots.sql")
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, upcomingSlotsSQL, at.Format("2006-01-02 15:04:05"), limit)
	if err != nil {
		return nil, fmt.Errorf("query upcoming slots: %w", err)
	}
	defer rows.Close()

	var slots []Slot
	for rows.Next() {
		var slot Slot
		if err := rows.Scan(
			&slot.Type,
			&slot.Period,
			&slot.Start,
			&slot.End,
			&slot.DurationMinutes,
		); err != nil {
			return nil, fmt.Errorf("scan upcoming slot: %w", err)
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate upcoming slots: %w", err)
	}

	return slots, nil
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
