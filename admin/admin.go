package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/chinmayyy01/ghostplay/storage"
)

type sessionSummary struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	RespStatus int       `json:"resp_status"`
	DurationMs int64     `json:"duration_ms"`
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next(w, r)
	}
}

func listSessionsHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		records, err := store.All()
		if err != nil {
			http.Error(w, "failed to read sessions", http.StatusInternalServerError)
			log.Printf("ADMIN ERROR listing sessions: %v", err)
			return
		}

		summaries := make([]sessionSummary, 0, len(records))
		for _, rec := range records {
			summaries = append(summaries, sessionSummary{
				ID:         rec.ID,
				Timestamp:  rec.Timestamp,
				Method:     rec.Method,
				Path:       rec.Path,
				RespStatus: rec.RespStatus,
				DurationMs: rec.DurationMs,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summaries)
	}
}

func getSessionHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		record, err := store.Get(id)
		if err != nil {
			http.Error(w, "failed to read sessions", http.StatusInternalServerError)
			log.Printf("ADMIN ERROR getting session %s: %v", id, err)
			return
		}
		if record == nil {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	}
}

func NewMux(store *storage.Store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/sessions", withCORS(listSessionsHandler(store)))
	mux.HandleFunc("GET /api/sessions/{id}", withCORS(getSessionHandler(store)))

	return mux
}