package http

import (
	"errors"
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
		Ingest(w, jobQueue, r)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		HealthCheck(w, r)
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
