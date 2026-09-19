package starter

import (
	"log"
	"time"

	"github.com/Shanu7002/ingestionEdge/internal/domain"
	"github.com/Shanu7002/ingestionEdge/internal/env"
	amqp "github.com/rabbitmq/amqp091-go"
)

// handle connections

func handleExchangeConnection(workerID int, ch *amqp.Channel) {
	err := ch.ExchangeDeclare(
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
}

func handleRabbitConnection() *amqp.Connection {
	log.Println("Connecting to RabbitMQ...")

	rabbitUrl := env.GetString("RABBITMQ_URL", "amqp://secretUser:secretPassword@localhost:5672/")

	var conn *amqp.Connection
	var err error

	for attempt := 1; attempt <= 5; attempt++ {
		conn, err = amqp.Dial(rabbitUrl)
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

	return conn
}

// handle creations

func HandleQueueCreation(capacity int) chan domain.IngestionPayload {
	return make(chan domain.IngestionPayload, capacity)
}

func handleChannelCreation(workerID int, conn *amqp.Connection) (*amqp.Channel, error) {
	return conn.Channel()
}
