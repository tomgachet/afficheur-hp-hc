package web_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appdb "afficheur-hp-hc/internal/db"
	"afficheur-hp-hc/internal/web"
)

func TestStatusEndpoint(t *testing.T) {
	ctx := context.Background()
	conn, err := appdb.Open(ctx, "", false)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/status?at=2026-06-17%2014:30:00", nil)
	rec := httptest.NewRecorder()

	web.NewServer(conn).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		CurrentType   string `json:"currentType"`
		CurrentPeriod string `json:"currentPeriod"`
		Remaining     string `json:"remaining"`
		NextType      string `json:"nextType"`
		UpcomingSlots []struct {
			Type     string `json:"type"`
			Duration string `json:"duration"`
		} `json:"upcomingSlots"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.CurrentType != "HC" {
		t.Fatalf("currentType = %q, want HC", body.CurrentType)
	}
	if body.CurrentPeriod != "Ete" {
		t.Fatalf("currentPeriod = %q, want Ete", body.CurrentPeriod)
	}
	if body.Remaining != "02h30" {
		t.Fatalf("remaining = %q, want 02h30", body.Remaining)
	}
	if body.NextType != "HP" {
		t.Fatalf("nextType = %q, want HP", body.NextType)
	}
	if len(body.UpcomingSlots) != 3 {
		t.Fatalf("upcomingSlots len = %d, want 3", len(body.UpcomingSlots))
	}
	if body.UpcomingSlots[0].Type != "HP" {
		t.Fatalf("first upcoming slot type = %q, want HP", body.UpcomingSlots[0].Type)
	}
}

func TestStatusEndpointRejectsInvalidAt(t *testing.T) {
	ctx := context.Background()
	conn, err := appdb.Open(ctx, "", false)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/status?at=bad", nil)
	rec := httptest.NewRecorder()

	web.NewServer(conn).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
