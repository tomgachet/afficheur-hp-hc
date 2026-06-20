package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	appdb "afficheur-hp-hc/internal/db"
	"afficheur-hp-hc/internal/hphc"
)

func main() {
	var dbPath string
	var atValue string
	var reloadReference bool
	var noColor bool
	flag.StringVar(&dbPath, "db", "conso_elec.duckdb", "chemin de la base DuckDB")
	flag.StringVar(&atValue, "at", "", "horodatage de test au format 2006-01-02 15:04:05")
	flag.BoolVar(&reloadReference, "reload-ref", false, "recharge ressources/ref_time_slot.csv dans DuckDB")
	flag.BoolVar(&noColor, "no-color", false, "desactive les couleurs ANSI")
	flag.Parse()

	at := time.Now()
	if atValue != "" {
		parsed, err := time.ParseInLocation("2006-01-02 15:04:05", atValue, time.Local)
		if err != nil {
			log.Fatalf("horodatage invalide: %v", err)
		}
		at = parsed
	}

	ctx := context.Background()
	conn, err := appdb.Open(ctx, dbPath, reloadReference)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	status, err := hphc.CurrentSlot(ctx, conn, at)
	if err != nil {
		log.Fatal(err)
	}

	colorsEnabled = !noColor

	fmt.Println(title(appName))
	fmt.Printf("%s %s\n", label("Date :"), value(formatDateTimeFR(status.Timestamp)))
	fmt.Printf("%s %s\n", label("Maintenant :"), periodType(status.CurrentType))
	fmt.Printf("%s %s\n", label("Periode :"), value(status.CurrentPeriod))
	fmt.Printf("%s %s\n", label("Debut :"), value(formatDateTimeFR(status.CurrentStart)))
	fmt.Printf("%s %s\n", label("Fin :"), value(formatDateTimeFR(status.CurrentEnd)))
	fmt.Printf("%s %s\n", label("Temps restant :"), strong(formatMinutes(status.RemainingMinutes)))
	fmt.Printf(
		"%s %s de %s a %s\n",
		label("Prochaine plage :"),
		periodType(status.NextType),
		value(formatDateTimeFR(status.NextStart)),
		value(formatDateTimeFR(status.NextEnd)),
	)
}

const (
	appName    = "afficheur-hp-hc"
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
)

var colorsEnabled = true

func color(code, text string) string {
	if !colorsEnabled {
		return text
	}
	return code + text + ansiReset
}

func title(text string) string {
	return color(ansiBold+ansiCyan, text)
}

func label(text string) string {
	return color(ansiDim, text)
}

func value(text string) string {
	return color(ansiCyan, text)
}

func strong(text string) string {
	return color(ansiBold, text)
}

func periodType(text string) string {
	switch text {
	case "HC":
		return color(ansiBold+ansiGreen, text)
	case "HP":
		return color(ansiBold+ansiYellow, text)
	default:
		return strong(text)
	}
}

func formatDateTimeFR(value time.Time) string {
	return fmt.Sprintf("%s %s", weekdayFR(value.Weekday()), value.Format("02/01/2006 15:04:05"))
}

func formatMinutes(minutes int) string {
	if minutes < 0 {
		minutes = 0
	}
	return fmt.Sprintf("%02dh%02d", minutes/60, minutes%60)
}

func weekdayFR(day time.Weekday) string {
	switch day {
	case time.Monday:
		return "lundi"
	case time.Tuesday:
		return "mardi"
	case time.Wednesday:
		return "mercredi"
	case time.Thursday:
		return "jeudi"
	case time.Friday:
		return "vendredi"
	case time.Saturday:
		return "samedi"
	case time.Sunday:
		return "dimanche"
	default:
		return ""
	}
}
