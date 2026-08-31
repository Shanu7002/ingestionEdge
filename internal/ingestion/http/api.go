package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Shanu7002/ingestionEdge/internal/domain"
)

type InternalIngestionPayload struct {
	Sender   string `json:"sender"`
	Alert    string `json:"alert"`
	Priority int    `json:"priority"`
}

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
	}

	sender := "null"
	if rawSender, ok := fields["sender"]; ok {
		if err := json.Unmarshal(rawSender, &sender); err != nil {
			http.Error(w, "Invalid sender", http.StatusBadRequest)
			return
		}
	}

	alert := ""
	if rawAlert, ok := fields["alert"]; ok {
		if err := json.Unmarshal(rawAlert, &alert); err != nil {
			http.Error(w, "Invalid alert", http.StatusBadRequest)
			return
		}
	}

	data, err := json.Marshal(InternalIngestionPayload{
		Sender:   sender,
		Alert:    alert,
		Priority: priority,
	})
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	payload := domain.IngestionPayload{
		Priority:   priority,
		Sender:     sender,
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
