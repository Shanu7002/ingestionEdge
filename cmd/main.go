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
	jobQueue := make(chan domain.IngestionPayload, 10000)

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
			for payload := range jobQueue {
				// TODO: serialize to protobuf & publish to RabbitMQ exchange

				// just simulate to see the UDP server happen
				// payload arrived intact
				log.Printf("Worker %d processed | Priority: %d | Time: %d | Data: %s",
					workerID, payload.Priority, payload.IngestedAt, string(payload.Data))

				// simulate latency
				// time.Sleep(50 * time.Millisecond) // for stress test
				time.Sleep(10 * time.Millisecond)

				if payload.Release != nil {
					payload.Release()
				}
			}
		}(i)
	}

	udpConn, err := ingestion.StartUDPServer(":8125", jobQueue, &wg, udpBufferPool)
	if err != nil {
		log.Fatalf("Fatal UDP error: %v", err)
	}

	httpServer := ingestion.StartHTTPServer(":8080", jobQueue, &wg)

	grpcServer, err := ingestion.StartGRPCServer(":9090", jobQueue, &wg)
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

	close(jobQueue)

	wg.Wait()
	log.Println("Edge API terminated securely. Zero data loss on critical queues.")
}
