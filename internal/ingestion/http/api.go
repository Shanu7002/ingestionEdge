package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Shanu7002/ingestionEdge/internal/domain"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed.\n Just GET is accepted here.", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"alive": true}`)
}

// TODO: save the alert in a database
// create an endpoint to get this alerts
func Ingest(w http.ResponseWriter, jobQueue chan<- domain.IngestionPayload, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	priority := 1

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if rawPriority, ok := fields["priority"]; ok {
		if err := json.Unmarshal(rawPriority, &priority); err != nil {
			http.Error(w, "Invalid priority", http.StatusBadRequest)
			return
		}
		delete(fields, "priority")
	}

	data, err := json.Marshal(fields)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	payload := domain.IngestionPayload{
		Priority:   priority,
		IngestedAt: time.Now().UnixNano(),
		Data:       data,
		Release:    nil,
	}

	select {
	case jobQueue <- payload:
		w.WriteHeader(http.StatusAccepted)
	case <-time.After(50 * time.Millisecond):
		http.Error(w, "Edge overloaded", http.StatusTooManyRequests)
	}
}
