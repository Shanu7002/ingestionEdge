package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shanu7002/ingestionEdge/cmd/starter"
	"github.com/Shanu7002/ingestionEdge/internal/ingestion"
	ingestionhttp "github.com/Shanu7002/ingestionEdge/internal/ingestion/http"
)

func main() {
	if err := run(); err != nil {
		// At this point, all defers inside run() have safely executed.
		// It is now safe to exit the process.
		log.Fatalf("Fatal error: %v", err)
	}
}

// ?IDEA: reuse the database to save metrics and alerts when rabbit its off and server up, we can reprocess this -> timestamp should come in payload.
func run() error {
	log.Println("Initializing...")

	var wg sync.WaitGroup
	metricsQueue := starter.HandleQueueCreation(10000)
	alertsQueue := starter.HandleQueueCreation(1000)

	udpBufferPool := &sync.Pool{
		New: func() interface{} {
			b := make([]byte, 1024)
			return &b
		},
	}
	var err error

	err = starter.HandleWorkerProcess(&wg, alertsQueue, metricsQueue)
	if err != nil {
		return fmt.Errorf("Fatal Worker Process error: %v", err)
	}

	udpConn, err := ingestion.StartUDPServer(":8125", metricsQueue, &wg, udpBufferPool)
	if err != nil {
		return fmt.Errorf("Fatal UDP error: %v", err)
	}

	httpServer := ingestionhttp.StartHTTPServer(":8080", alertsQueue, &wg)

	grpcServer, err := ingestion.StartGRPCServer(":9090", alertsQueue, &wg)
	if err != nil {
		return fmt.Errorf("Fatal gRPC error: %v", err)
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
		return fmt.Errorf("HTTP shutdown error: %v", err)
	}

	close(alertsQueue)
	close(metricsQueue)

	wg.Wait()
	return fmt.Errorf("Edge API terminated securely. Zero data loss on critical queues.")
}
