package ingestion

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/Shanu7002/ingestionEdge/internal/domain"
)

func StartUDPServer(
	addr string,
	jobQueue chan<- domain.IngestionPayload,
	wg *sync.WaitGroup,
	pool *sync.Pool,
) (*net.UDPConn, error) {

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("UDP Server listening on %s", addr)

		for {
			bufPtr := pool.Get().(*[]byte)
			n, _, err := udpConn.ReadFromUDP(*bufPtr)

			timestamp := time.Now().UnixNano()

			if err != nil {
				return
			}

			payload := domain.IngestionPayload{
				Priority:   0,
				IngestedAt: timestamp,
				Data:       (*bufPtr)[:n],
				Release: func() {
					pool.Put(bufPtr)
				},
			}

			select {
			case jobQueue <- payload:
			default:
				// queue full
				payload.Release()
				log.Println("Warn: UDP load shedding triggered. Metric dropped.")
			}
		}
	}()

	return udpConn, nil
}
