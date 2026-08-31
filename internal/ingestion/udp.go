package ingestion

import (
	"bytes"
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
			n, senderAddr, err := udpConn.ReadFromUDP(*bufPtr)
			if err != nil {
				return
			}

			timestamp := time.Now().UnixNano()

			data := (*bufPtr)[:n]

			parts := bytes.SplitN(data, []byte("|"), 2)

			// sender is actually IP:PORT so you can match the IP:port with a name to replace
			sender := senderAddr.String()
			if len(parts) == 2 {
				sender = string(parts[0])
				data = parts[1]
			}

			payload := domain.IngestionPayload{
				Priority:   0,
				Sender:     sender,
				IngestedAt: timestamp,
				Data:       data,
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
