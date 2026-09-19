package starter

import (
	"log"
	"runtime"
	"sync"

	"github.com/Shanu7002/ingestionEdge/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

func HandleWorkerProcess(
	wg *sync.WaitGroup,
	alertsQueue chan domain.IngestionPayload,
	metricsQueue chan domain.IngestionPayload,
) error {
	numWorkers := handleWorkerSpawn()
	log.Printf("Spawning %d RabbitMQ publisher workers", numWorkers)

	conn := handleRabbitConnection()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go handleRunWorker(
			wg,
			i,
			conn,
			alertsQueue,
			metricsQueue,
		)
	}

	return nil
}

func handleRunWorker(
	wg *sync.WaitGroup,
	workerID int,
	conn *amqp.Connection,
	alertsQueue chan domain.IngestionPayload,
	metricsQueue chan domain.IngestionPayload,
) {
	defer wg.Done()

	ch, err := handleChannelCreation(workerID, conn)
	if err != nil {
		log.Printf("Worker %d failed to open channel: %v", workerID, err)
		return
	}
	defer ch.Close()

	handleExchangeConnection(workerID, ch)
	defer conn.Close()

	runWorkerLoop(
		workerID,
		ch,
		alertsQueue,
		metricsQueue,
	)

}

func runWorkerLoop(
	workerID int,
	ch *amqp.Channel,
	alertsQueue chan domain.IngestionPayload,
	metricsQueue chan domain.IngestionPayload,
) {
	for {
		payload, ok := receivePayload(alertsQueue, metricsQueue)
		if !ok {
			// TODO: should go to an error queue
			return
		}

		processPayload(workerID, payload, ch)
	}
}

func handleWorkerSpawn() int {
	return runtime.NumCPU() * 2
}
