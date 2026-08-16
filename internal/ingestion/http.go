package ingestion

import (
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Shanu7002/ingestionEdge/internal/domain"
)

func StartHTTPServer(
	addr string,
	jobQueue chan<- domain.IngestionPayload,
	wg *sync.WaitGroup,
) *http.Server {

	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
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

		timestamp := time.Now().UnixNano()

		payload := domain.IngestionPayload{
			Priority:   1,
			IngestedAt: timestamp,
			Data:       body,
			Release:    nil,
		}

		select {
		case jobQueue <- payload:
			w.WriteHeader(http.StatusAccepted)
		case <-time.After(50 * time.Millisecond):
			http.Error(w, "Edge overloaded", http.StatusTooManyRequests)
		}
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("HTTP Server listening on %s (Fallback Alerts)", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server crashed: %v", err)
		}
	}()

	return server
}
