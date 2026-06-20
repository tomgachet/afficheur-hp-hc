package hphc_test

import (
	"context"
	"testing"
	"time"

	appdb "afficheur-hp-hc/internal/db"
	"afficheur-hp-hc/internal/hphc"
)

func TestCurrentSlot(t *testing.T) {
	ctx := context.Background()
	conn, err := appdb.Open(ctx, "", false)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tests := []struct {
		name          string
		at            string
		currentType   string
		currentPeriod string
		currentEnd    string
		nextType      string
		nextStart     string
		nextEnd       string
	}{
		{
			name:          "lundi hiver dans la nuit",
			at:            "2026-01-05 03:00:00",
			currentType:   "HC",
			currentPeriod: "Hiver",
			currentEnd:    "2026-01-05 07:00:00",
			nextType:      "HP",
		},
		{
			name:          "lundi hiver debut HP",
			at:            "2026-01-05 07:00:00",
			currentType:   "HP",
			currentPeriod: "Hiver",
			currentEnd:    "2026-01-05 13:00:00",
			nextType:      "HC",
		},
		{
			name:          "lundi hiver debut HC midi",
			at:            "2026-01-05 13:00:00",
			currentType:   "HC",
			currentPeriod: "Hiver",
			currentEnd:    "2026-01-05 16:00:00",
			nextType:      "HP",
		},
		{
			name:          "lundi hiver debut HP apres midi",
			at:            "2026-01-05 16:00:00",
			currentType:   "HP",
			currentPeriod: "Hiver",
			currentEnd:    "2026-01-06 00:00:00",
			nextType:      "HC",
		},
		{
			name:          "samedi et dimanche colles en HC",
			at:            "2026-01-10 12:00:00",
			currentType:   "HC",
			currentPeriod: "Hiver",
			currentEnd:    "2026-01-12 07:00:00",
			nextType:      "HP",
			nextStart:     "2026-01-12 07:00:00",
			nextEnd:       "2026-01-12 13:00:00",
		},
		{
			name:          "dimanche colle avec lundi matin HC",
			at:            "2026-01-11 12:00:00",
			currentType:   "HC",
			currentPeriod: "Hiver",
			currentEnd:    "2026-01-12 07:00:00",
			nextType:      "HP",
			nextStart:     "2026-01-12 07:00:00",
			nextEnd:       "2026-01-12 13:00:00",
		},
		{
			name:          "borne avant 07h",
			at:            "2026-01-05 06:59:30",
			currentType:   "HC",
			currentPeriod: "Hiver",
			currentEnd:    "2026-01-05 07:00:00",
			nextType:      "HP",
		},
		{
			name:          "ete",
			at:            "2026-06-17 14:30:00",
			currentType:   "HC",
			currentPeriod: "Ete",
			currentEnd:    "2026-06-17 17:00:00",
			nextType:      "HP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at := mustParseTime(t, tt.at)
			status, err := hphc.CurrentSlot(ctx, conn, at)
			if err != nil {
				t.Fatal(err)
			}

			if status.CurrentType != tt.currentType {
				t.Fatalf("CurrentType = %q, want %q", status.CurrentType, tt.currentType)
			}
			if status.CurrentPeriod != tt.currentPeriod {
				t.Fatalf("CurrentPeriod = %q, want %q", status.CurrentPeriod, tt.currentPeriod)
			}
			if got := status.CurrentEnd.Format("2006-01-02 15:04:05"); got != tt.currentEnd {
				t.Fatalf("CurrentEnd = %s, want %s", got, tt.currentEnd)
			}
			if status.NextType != tt.nextType {
				t.Fatalf("NextType = %q, want %q", status.NextType, tt.nextType)
			}
			if tt.nextStart != "" {
				if got := status.NextStart.Format("2006-01-02 15:04:05"); got != tt.nextStart {
					t.Fatalf("NextStart = %s, want %s", got, tt.nextStart)
				}
			}
			if tt.nextEnd != "" {
				if got := status.NextEnd.Format("2006-01-02 15:04:05"); got != tt.nextEnd {
					t.Fatalf("NextEnd = %s, want %s", got, tt.nextEnd)
				}
			}
		})
	}
}

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
