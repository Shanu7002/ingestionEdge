package starter

import (
	"context"
	"log"
	"time"

	"github.com/Shanu7002/ingestionEdge/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

func receivePayload(
	alertsQueue chan domain.IngestionPayload,
	metricsQueue chan domain.IngestionPayload,
) (domain.IngestionPayload, bool) {
	select {
	case payload, ok := <-alertsQueue:
		return payload, ok
	default:
	}

	select {
	case payloadd, ok := <-alertsQueue:
		return payloadd, ok

	case payload, ok := <-metricsQueue:
		return payload, ok
	}
}

func processPayload(workerID int, payload domain.IngestionPayload, ch *amqp.Channel) {
	// debugging
	// log.Printf("Worker %d | Sender: %s | Priority: %d | Time: %d | Data: %s", workerID, payload.Sender, payload.Priority, payload.IngestedAt, string(payload.Data))

	routingKey := "route.metric"
	if payload.Priority == 1 {
		// TODO: insert alert in database
		routingKey = "route.alert"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := ch.PublishWithContext(
		ctx,
		"telemetry.exchange", // exchange
		routingKey,           // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,            // save to disk (for alerts)
			ContentType:  "application/octet-stream", // Binary Protobuf
			Timestamp:    time.Unix(0, payload.IngestedAt),
			Body:         payload.Data,
		},
	)

	if err != nil {
		log.Printf("Worker %d | Failed to publish payload: %v", workerID, err)
		// TODO: insert metric fallback here for retry later
		// TODO: since its an error and could be an alerts, I'll need a reprocess alert idea flux
	}

	if payload.Release != nil {
		payload.Release()
	}
}
