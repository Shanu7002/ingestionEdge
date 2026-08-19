package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/Shanu7002/ingestionEdge/internal/domain"
	"github.com/Shanu7002/ingestionEdge/internal/ingestion"
)

func main() {
	log.Println("Initializing Edge API Orchestrator...")

	var wg sync.WaitGroup
	metricsQueue := make(chan domain.IngestionPayload, 10000)
	alertsQueue := make(chan domain.IngestionPayload, 1000)

	udpBufferPool := &sync.Pool{
		New: func() interface{} {
			b := make([]byte, 1024)
			return &b
		},
	}

	numWorkers := runtime.NumCPU() * 2
	log.Printf("Spawning %d RabbitMQ publisher workers", numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				var payload domain.IngestionPayload
				var ok bool

				// the worker checks the alertsQueue first.
				select {
				case payload, ok = <-alertsQueue:
					if !ok {
						return
					} // channel closed during shutdown
					processPayload(workerID, payload)
					continue // back to see if there are more alerts
				default:
				}

				select {
				case payload, ok = <-alertsQueue:
					if !ok {
						return
					}
					processPayload(workerID, payload)
				case payload, ok = <-metricsQueue:
					if !ok {
						return
					}
					processPayload(workerID, payload)
				}
			}
		}(i)
	}

	udpConn, err := ingestion.StartUDPServer(":8125", metricsQueue, &wg, udpBufferPool)
	if err != nil {
		log.Fatalf("Fatal UDP error: %v", err)
	}

	httpServer := ingestion.StartHTTPServer(":8080", alertsQueue, &wg)

	grpcServer, err := ingestion.StartGRPCServer(":9090", alertsQueue, &wg)
	if err != nil {
		log.Fatalf("Fatal gRPC error: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("SIGTERM received. Halting network ingestion...")
	udpConn.Close()
	grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	close(alertsQueue)
	close(metricsQueue)

	wg.Wait()
	log.Println("Edge API terminated securely. Zero data loss on critical queues.")
}

// Helper function to keep the loop clean
func processPayload(workerID int, payload domain.IngestionPayload) {
	// debugging
	// log.Printf("Worker %d | Priority: %d | Time: %d | Data: %s", workerID, payload.Priority, payload.IngestedAt, string(payload.Data))

	// TODO: Publish to RabbitMQ

	// Simulate RabbitMQ I/O Latency
	time.Sleep(10 * time.Millisecond)

	if payload.Release != nil {
		payload.Release()
	}
}
