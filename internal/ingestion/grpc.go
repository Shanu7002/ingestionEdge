package ingestion

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/Shanu7002/ingestionEdge/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func StartGRPCServer(
	addr string,
	jobQueue chan<- domain.IngestionPayload, // Will be passed to the RPC handlers
	wg *sync.WaitGroup,
) (*grpc.Server, error) {

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			MaxConnectionAge:  30 * time.Minute,
			Time:              1 * time.Minute,
			Timeout:           20 * time.Second,
		}),
		grpc.MaxConcurrentStreams(1000),
	)

	// TODO: register your gRPC service here once protobufs are compiled.

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("gRPC Server listening on %s", addr)
		if err := server.Serve(listener); err != nil {
			log.Fatalf("gRPC server crashed: %v", err)
		}
	}()

	return server, nil
}
