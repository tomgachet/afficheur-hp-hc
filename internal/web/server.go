package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"afficheur-hp-hc/internal/hphc"
)

type Server struct {
	db *sql.DB
}

func NewServer(db *sql.DB) http.Handler {
	server := &Server{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.index)
	mux.HandleFunc("/api/status", server.status)
	return mux
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	at, err := parseAt(r.URL.Query().Get("at"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	status, err := hphc.CurrentSlot(r.Context(), s.db, at)
	if err != nil {
		code := http.StatusInternalServerError
		if err == hphc.ErrNoSlot {
			code = http.StatusNotFound
		}
		http.Error(w, err.Error(), code)
		return
	}

	upcomingSlots, err := hphc.UpcomingSlots(r.Context(), s.db, at, 3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(newStatusResponse(status, upcomingSlots)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseAt(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Now(), nil
	}

	return time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
}

type statusResponse struct {
	Timestamp              string         `json:"timestamp"`
	CurrentType            string         `json:"currentType"`
	CurrentPeriod          string         `json:"currentPeriod"`
	CurrentStart           string         `json:"currentStart"`
	CurrentEnd             string         `json:"currentEnd"`
	CurrentStartTime       string         `json:"currentStartTime"`
	CurrentEndTime         string         `json:"currentEndTime"`
	CurrentStartWeekday    string         `json:"currentStartWeekday"`
	CurrentEndWeekday      string         `json:"currentEndWeekday"`
	CurrentDurationMinutes int            `json:"currentDurationMinutes"`
	ElapsedMinutes         int            `json:"elapsedMinutes"`
	ProgressPercent        int            `json:"progressPercent"`
	RemainingMinutes       int            `json:"remainingMinutes"`
	RemainingHours         int            `json:"remainingHours"`
	RemainingPart          int            `json:"remainingPart"`
	Remaining              string         `json:"remaining"`
	NextType               string         `json:"nextType"`
	NextPeriod             string         `json:"nextPeriod"`
	NextStart              string         `json:"nextStart"`
	NextEnd                string         `json:"nextEnd"`
	NextStartTime          string         `json:"nextStartTime"`
	NextEndTime            string         `json:"nextEndTime"`
	NextStartWeekday       string         `json:"nextStartWeekday"`
	NextEndWeekday         string         `json:"nextEndWeekday"`
	NextDuration           string         `json:"nextDuration"`
	UpcomingSlots          []slotResponse `json:"upcomingSlots"`
}

type slotResponse struct {
	Type         string `json:"type"`
	Period       string `json:"period"`
	Start        string `json:"start"`
	End          string `json:"end"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	StartWeekday string `json:"startWeekday"`
	EndWeekday   string `json:"endWeekday"`
	Duration     string `json:"duration"`
}

func newStatusResponse(status hphc.SlotStatus, upcomingSlots []hphc.Slot) statusResponse {
	remaining := cleanMinutes(status.RemainingMinutes)
	duration := int(status.CurrentEnd.Sub(status.CurrentStart).Minutes())
	elapsed := int(status.Timestamp.Sub(status.CurrentStart).Minutes())
	progress := 0
	if duration > 0 {
		progress = elapsed * 100 / duration
	}

	return statusResponse{
		Timestamp:              formatDateTimeFR(status.Timestamp),
		CurrentType:            status.CurrentType,
		CurrentPeriod:          status.CurrentPeriod,
		CurrentStart:           formatDateTimeFR(status.CurrentStart),
		CurrentEnd:             formatDateTimeFR(status.CurrentEnd),
		CurrentStartTime:       formatTime(status.CurrentStart),
		CurrentEndTime:         formatTime(status.CurrentEnd),
		CurrentStartWeekday:    formatDayLabel(status.CurrentStart),
		CurrentEndWeekday:      formatDayLabel(status.CurrentEnd),
		CurrentDurationMinutes: cleanMinutes(duration),
		ElapsedMinutes:         cleanMinutes(elapsed),
		ProgressPercent:        clamp(progress, 0, 100),
		RemainingMinutes:       remaining,
		RemainingHours:         remaining / 60,
		RemainingPart:          remaining % 60,
		Remaining:              formatMinutes(remaining),
		NextType:               status.NextType,
		NextPeriod:             status.NextPeriod,
		NextStart:              formatDateTimeFR(status.NextStart),
		NextEnd:                formatDateTimeFR(status.NextEnd),
		NextStartTime:          formatTime(status.NextStart),
		NextEndTime:            formatTime(status.NextEnd),
		NextStartWeekday:       formatDayLabel(status.NextStart),
		NextEndWeekday:         formatDayLabel(status.NextEnd),
		NextDuration:           formatMinutes(status.NextDurationMinutes),
		UpcomingSlots:          newSlotResponses(upcomingSlots),
	}
}

func newSlotResponses(slots []hphc.Slot) []slotResponse {
	responses := make([]slotResponse, 0, len(slots))
	for _, slot := range slots {
		responses = append(responses, slotResponse{
			Type:         slot.Type,
			Period:       slot.Period,
			Start:        formatDateTimeFR(slot.Start),
			End:          formatDateTimeFR(slot.End),
			StartTime:    formatTime(slot.Start),
			EndTime:      formatTime(slot.End),
			StartWeekday: formatDayLabel(slot.Start),
			EndWeekday:   formatDayLabel(slot.End),
			Duration:     formatMinutes(slot.DurationMinutes),
		})
	}
	return responses
}

func formatDateTimeFR(value time.Time) string {
	return fmt.Sprintf("%s %s", weekdayFR(value.Weekday()), value.Format("02/01/2006 15:04:05"))
}

func formatTime(value time.Time) string {
	return value.Format("15:04")
}

func formatDayLabel(value time.Time) string {
	return fmt.Sprintf("%s %02d", weekdayFR(value.Weekday()), value.Day())
}

func formatMinutes(minutes int) string {
	minutes = cleanMinutes(minutes)
	return fmt.Sprintf("%02dh%02d", minutes/60, minutes%60)
}

func cleanMinutes(minutes int) int {
	if minutes < 0 {
		return 0
	}
	return minutes
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
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

var indexTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="fr">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>afficheur-hp-hc</title>
  <style>
    :root {
      color-scheme: light dark;
      --bg: #f6f7f3;
      --panel: #ffffff;
      --text: #1c2420;
      --muted: #657169;
      --line: #dfe4dd;
      --hc: #067a46;
      --hp: #a05a00;
      --error: #b42318;
    }

    @media (prefers-color-scheme: dark) {
      :root {
        --bg: #111511;
        --panel: #1b211c;
        --text: #eff4ee;
        --muted: #aeb8af;
        --line: #30382f;
        --hc: #39d98a;
        --hp: #ffbf47;
        --error: #ff8a80;
      }
    }

    * { box-sizing: border-box; }

    body {
      margin: 0;
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 24px;
      background: var(--bg);
      color: var(--text);
      font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }

    main { width: min(100%, 620px); }

    header {
      display: flex;
      align-items: end;
      justify-content: space-between;
      gap: 16px;
      margin-bottom: 18px;
    }

    h1 {
      margin: 0;
      font-size: clamp(1.25rem, 3vw, 1.8rem);
      line-height: 1;
      letter-spacing: 0;
    }

    .top-status {
      display: grid;
      gap: 4px;
      justify-items: end;
      text-align: right;
    }

    .top-period {
      color: var(--text);
      font-size: clamp(1rem, 3vw, 1.15rem);
      font-weight: 850;
      white-space: nowrap;
    }

    .now {
      color: var(--muted);
      font-size: clamp(1.05rem, 3vw, 1.28rem);
      font-weight: 800;
      white-space: nowrap;
    }

    .panel {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 22px;
      box-shadow: 0 18px 45px rgb(0 0 0 / 8%);
    }

    .current {
      display: grid;
      gap: 10px;
      margin-bottom: 22px;
    }

    .headline {
      display: flex;
      align-items: baseline;
      justify-content: space-between;
      gap: 14px;
      flex-wrap: wrap;
    }

    .badge {
      width: max-content;
      min-width: 68px;
      padding: 6px 12px;
      border-radius: 999px;
      font-size: 1.1rem;
      font-weight: 800;
      text-align: center;
      border: 1px solid currentColor;
    }

    .badge.HC { color: var(--hc); }
    .badge.HP { color: var(--hp); }

    .remaining {
      display: flex;
      align-items: baseline;
      gap: 10px;
      flex-wrap: wrap;
    }

    .remaining-value {
      font-size: clamp(1.45rem, 5vw, 2.25rem);
      line-height: 1;
      font-weight: 850;
      letter-spacing: 0;
    }

    .remaining-label {
      color: var(--muted);
      font-size: clamp(1rem, 3vw, 1.3rem);
      font-weight: 800;
    }


    .timeline {
      display: grid;
      gap: 14px;
      margin: 8px 0 22px;
      padding: 18px 0;
      border-top: 1px solid var(--line);
      border-bottom: 1px solid var(--line);
    }

    .timeline-head,
    .timeline-labels {
      display: flex;
      justify-content: space-between;
      gap: 12px;
    }

    .timeline-title {
      color: var(--muted);
      font-weight: 800;
    }

    .timeline-range {
      display: grid;
      gap: 2px;
      font-weight: 750;
      text-align: right;
    }

    .range-days {
      color: var(--muted);
      font-size: 0.9rem;
      font-weight: 700;
    }

    .track {
      position: relative;
      height: 24px;
      overflow: hidden;
      border-radius: 999px;
      background: #e7ebe5;
      border: 1px solid var(--line);
    }

    @media (prefers-color-scheme: dark) {
      .track { background: #30382f; }
    }

    .track-fill {
      width: 0%;
      height: 100%;
      background: var(--hc);
      transition: width 250ms ease, background-color 250ms ease;
    }

    .track-fill.HP { background: var(--hp); }
    .track-fill.HC { background: var(--hc); }

    .marker {
      position: absolute;
      top: 50%;
      left: 0%;
      width: 18px;
      height: 18px;
      border-radius: 999px;
      background: var(--panel);
      border: 3px solid var(--text);
      transform: translate(-50%, -50%);
      transition: left 250ms ease;
      box-shadow: 0 1px 4px rgb(0 0 0 / 22%);
    }

    .timeline-labels {
      color: var(--muted);
      font-size: 0.95rem;
      font-weight: 750;
    }

    .timeline-next {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      padding: 12px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #f0f3ee;
    }

    @media (prefers-color-scheme: dark) {
      .timeline-next { background: #222a23; }
    }

    .timeline-next-info {
      display: grid;
      gap: 2px;
      min-width: 0;
    }

    .timeline-next-title {
      color: var(--muted);
      font-size: 0.9rem;
      font-weight: 800;
    }

    .timeline-next-days {
      color: var(--muted);
      font-size: 0.9rem;
      font-weight: 650;
    }

    .timeline-next-range {
      overflow-wrap: anywhere;
      font-weight: 750;
    }


    .upcoming {
      margin-top: 14px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #f8faf7;
    }

    @media (prefers-color-scheme: dark) {
      .upcoming { background: #171d18; }
    }

    .upcoming summary {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      padding: 12px;
      cursor: pointer;
      color: var(--text);
      font-weight: 800;
      list-style: none;
    }

    .upcoming summary::-webkit-details-marker { display: none; }

    .upcoming summary::after {
      content: "+";
      color: var(--muted);
      font-size: 1.2rem;
      line-height: 1;
    }

    .upcoming[open] summary::after { content: "-"; }

    .upcoming-list {
      display: grid;
      gap: 0;
      padding: 0 12px 12px;
    }

    .upcoming-slot {
      display: grid;
      grid-template-columns: auto 1fr;
      gap: 10px;
      align-items: start;
      padding: 12px 0;
      border-top: 1px solid var(--line);
    }

    .upcoming-slot-main {
      display: grid;
      gap: 3px;
      min-width: 0;
    }

    .upcoming-slot-days {
      color: var(--muted);
      font-size: 0.9rem;
      font-weight: 700;
    }

    .upcoming-slot-range {
      font-weight: 800;
      overflow-wrap: anywhere;
    }

    .upcoming-slot-meta {
      color: var(--muted);
      font-size: 0.9rem;
      font-weight: 650;
    }

    .status {
      min-height: 1.3em;
      margin-top: 14px;
      color: var(--muted);
      font-size: 0.95rem;
    }

    .status.error { color: var(--error); }
  </style>
</head>
<body>
  <main>
    <header>
      <h1>afficheur-hp-hc</h1>
      <div class="top-status">
        <div id="top-period" class="top-period">--</div>
        <div id="now" class="now">--</div>
      </div>
    </header>

    <section class="panel" aria-live="polite">
      <div class="current">
        <div class="headline">
          <span id="current-type" class="badge">--</span>
          <div class="remaining" aria-label="Temps restant">
            <span id="remaining-value" class="remaining-value">--h--</span>
            <span class="remaining-label">restantes</span>
          </div>
        </div>
      </div>

      <div class="timeline" aria-label="Timeline des plages tarifaires">
        <div class="timeline-head">
          <span class="timeline-title">Plage en cours</span>
          <span class="timeline-range">
            <span id="timeline-current-days" class="range-days">--</span>
            <span id="timeline-current-range">--</span>
          </span>
        </div>
        <div class="track">
          <div id="timeline-fill" class="track-fill"></div>
          <div id="timeline-marker" class="marker"></div>
        </div>
        <div class="timeline-labels">
          <span id="timeline-start">--:--</span>
          <span id="timeline-progress">--%</span>
          <span id="timeline-end">--:--</span>
        </div>
        <div class="timeline-next">
          <div class="timeline-next-info">
            <span class="timeline-next-title">Prochaine plage</span>
            <span id="timeline-next-days" class="timeline-next-days">--</span>
            <span id="timeline-next-range" class="timeline-next-range">--</span>
          </div>
          <span id="timeline-next-type" class="badge">--</span>
        </div>

        <details class="upcoming">
          <summary>Voir les 3 prochaines plages</summary>
          <div id="upcoming-list" class="upcoming-list"></div>
        </details>
      </div>



      <div id="status" class="status"></div>
    </section>
  </main>

  <script>
    const fields = {
      now: document.querySelector("#now"),
      topPeriod: document.querySelector("#top-period"),
      currentType: document.querySelector("#current-type"),
      remainingValue: document.querySelector("#remaining-value"),
      timelineCurrentDays: document.querySelector("#timeline-current-days"),
      timelineCurrentRange: document.querySelector("#timeline-current-range"),
      timelineFill: document.querySelector("#timeline-fill"),
      timelineMarker: document.querySelector("#timeline-marker"),
      timelineStart: document.querySelector("#timeline-start"),
      timelineProgress: document.querySelector("#timeline-progress"),
      timelineEnd: document.querySelector("#timeline-end"),
      timelineNextDays: document.querySelector("#timeline-next-days"),
      timelineNextRange: document.querySelector("#timeline-next-range"),
      timelineNextType: document.querySelector("#timeline-next-type"),
      upcomingList: document.querySelector("#upcoming-list"),
      status: document.querySelector("#status"),
    };

    function setBadge(element, value) {
      element.textContent = value || "--";
      element.classList.toggle("HC", value === "HC");
      element.classList.toggle("HP", value === "HP");
    }

    function formatRangeDays(startDay, endDay) {
      if (!startDay || !endDay) {
        return "--";
      }
      if (startDay === endDay) {
        return startDay;
      }
      return startDay + " - " + endDay;
    }

    function renderUpcoming(slots) {
      fields.upcomingList.replaceChildren(...(slots || []).map((slot) => {
        const row = document.createElement("div");
        row.className = "upcoming-slot";

        const badge = document.createElement("span");
        badge.className = "badge";
        setBadge(badge, slot.type);

        const main = document.createElement("div");
        main.className = "upcoming-slot-main";

        const days = document.createElement("span");
        days.className = "upcoming-slot-days";
        days.textContent = formatRangeDays(slot.startWeekday, slot.endWeekday);

        const range = document.createElement("span");
        range.className = "upcoming-slot-range";
        range.textContent = slot.startTime + " - " + slot.endTime;

        const meta = document.createElement("span");
        meta.className = "upcoming-slot-meta";
        meta.textContent = "Periode " + slot.period + " | " + slot.duration;

        main.append(days, range, meta);
        row.append(badge, main);
        return row;
      }));
    }

    function updateNow() {
      const now = new Date();
      const date = now.toLocaleDateString("fr-FR", {
        weekday: "long",
        day: "numeric",
        month: "long",
        year: "numeric",
      });
      const time = now.toLocaleTimeString("fr-FR", {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      });
      fields.now.textContent = date + " | " + time;
    }

    async function refresh() {
      try {
        const response = await fetch("/api/status", { cache: "no-store" });
        if (!response.ok) {
          throw new Error(await response.text());
        }

        const data = await response.json();
        setBadge(fields.currentType, data.currentType);
        setBadge(fields.timelineNextType, data.nextType);
        fields.topPeriod.textContent = "Periode " + data.currentPeriod;
        fields.remainingValue.textContent = data.remaining;
        fields.timelineCurrentDays.textContent = formatRangeDays(data.currentStartWeekday, data.currentEndWeekday);
        fields.timelineCurrentRange.textContent = data.currentStartTime + " - " + data.currentEndTime;
        fields.timelineStart.textContent = data.currentStartTime;
        fields.timelineEnd.textContent = data.currentEndTime;
        fields.timelineProgress.textContent = data.progressPercent + "%";
        fields.timelineFill.style.width = data.progressPercent + "%";
        fields.timelineMarker.style.left = data.progressPercent + "%";
        fields.timelineFill.classList.toggle("HC", data.currentType === "HC");
        fields.timelineFill.classList.toggle("HP", data.currentType === "HP");
        fields.timelineNextDays.textContent = formatRangeDays(data.nextStartWeekday, data.nextEndWeekday);
        fields.timelineNextRange.textContent = data.nextStartTime + " - " + data.nextEndTime + " (" + data.nextDuration + ")";
        renderUpcoming(data.upcomingSlots);
        fields.status.textContent = "Mis a jour a " + new Date().toLocaleTimeString("fr-FR");
        fields.status.classList.remove("error");
      } catch (error) {
        fields.status.textContent = error.message.trim() || "Erreur de mise a jour";
        fields.status.classList.add("error");
      }
    }

    updateNow();
    refresh();
    setInterval(updateNow, 1000);
    setInterval(refresh, 30000);
  </script>
</body>
</html>`))
