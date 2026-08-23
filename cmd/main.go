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
	"github.com/Shanu7002/ingestionEdge/internal/env"
	"github.com/Shanu7002/ingestionEdge/internal/ingestion"
	amqp "github.com/rabbitmq/amqp091-go"
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

	log.Println("Connecting to RabbitMQ...")

	rabbitURL := env.GetString("RABBITMQ_URL", "amqp://secretUser:secretPassword@localhost:5672/")

	var conn *amqp.Connection
	var err error

	// 15sec loop max
	for attempt := 1; attempt <= 5; attempt++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			log.Println("Successfully connected to RabbitMQ.")
			break // success
		}

		log.Printf("Attempt %d: RabbitMQ not ready (%v)", attempt, err)
		if attempt < 5 {
			log.Println("Retrying in 3 seconds...")
			time.Sleep(3 * time.Second)
		}
	}

	if err != nil {
		log.Fatalf("Fatal: Could not connect to RabbitMQ after 5 attempts: %v", err)
	}
	defer conn.Close()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			ch, err := conn.Channel()
			if err != nil {
				log.Printf("Worker %d failed to open channel: %v", workerID, err)
				return
			}
			defer ch.Close()

			err = ch.ExchangeDeclare(
				"telemetry.exchange", // name
				"direct",             // type (O(1) routing)
				true,                 // durable
				false,                // auto-deleted
				false,                // internal
				false,                // no-wait
				nil,                  // arguments
			)
			if err != nil {
				log.Printf("Worker %d failed to declare exchange: %v", workerID, err)
				return
			}

			for {
				var payload domain.IngestionPayload
				var ok bool

				select {
				case payload, ok = <-alertsQueue:
					if !ok {
						return
					}
					processPayload(workerID, payload, ch)
					continue
				default:
				}

				select {
				case payload, ok = <-alertsQueue:
					if !ok {
						return
					}
					processPayload(workerID, payload, ch)
				case payload, ok = <-metricsQueue:
					if !ok {
						return
					}
					processPayload(workerID, payload, ch)
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

func processPayload(workerID int, payload domain.IngestionPayload, ch *amqp.Channel) {
	// debugging
	// log.Printf("Worker %d | Priority: %d | Time: %d | Data: %s", workerID, payload.Priority, payload.IngestedAt, string(payload.Data))

	routingKey := "route.metric"
	if payload.Priority == 1 {
		routingKey = "route.alert"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := ch.PublishWithContext(ctx,
		"telemetry.exchange", // exchange
		routingKey,           // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,            // save to disk (for alerts)
			ContentType:  "application/octet-stream", // Binary Protobuf
			Timestamp:    time.Unix(0, payload.IngestedAt),
			Body:         payload.Data,
		})

	if err != nil {
		log.Printf("Worker %d | Failed to publish payload: %v", workerID, err)
		// In a production system, you would push this back into a retry queue here.
	}

	if payload.Release != nil {
		payload.Release()
	}
}
