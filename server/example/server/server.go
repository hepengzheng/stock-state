package main

import (
	"io"
	"log"
	"math/rand"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
	"github.com/hepengzheng/stock-state/pkg/logger"
	"github.com/hepengzheng/stock-state/pkg/logutil"
)

type MockServer struct {
	statepb.UnimplementedStateServer
}

func (ms *MockServer) GetStock(stream statepb.State_GetStockServer) error {
	defer logutil.LogPanic()
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			logger.Error("receive error", zap.Error(err))
			return err
		}

		r := rand.Intn(10)

		resp := &statepb.Response{
			RequestId: req.RequestId,
			Timestamp: req.Timestamp,
			Error:     statepb.Response_OK,
			Result:    int32(r),
		}

		if err = stream.Send(resp); err != nil {
			logger.Error("failed to send response", zap.Error(err),
				zap.String("key", req.Key), zap.String("request_id", resp.RequestId))
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register the ChatServiceServer
	statepb.RegisterStateServer(grpcServer, &MockServer{})

	log.Println("gRPC server is running on port 50051...")
	if err = grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
