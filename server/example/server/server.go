package main

import (
	"context"
	"math/rand"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
	"github.com/hepengzheng/stock-state/pkg/logger"
)

type MockServer struct {
	statepb.UnimplementedStateServer
}

func (ms *MockServer) GetStock(ctx context.Context, req *statepb.Request) (*statepb.Response, error) {
	resp := &statepb.Response{
		RequestId: req.RequestId,
		Timestamp: time.Now().UnixMicro(),
		Error:     statepb.Response_OK,
		Result:    int32(rand.Intn(10)),
	}
	return resp, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("failed to listen", zap.Error(err))
		return
	}
	grpcServer := grpc.NewServer()
	statepb.RegisterStateServer(grpcServer, &MockServer{})

	logger.Info("gRPC server is running on port 50051...")
	if err = grpcServer.Serve(listener); err != nil {
		logger.Error("failed to serve", zap.Error(err))
	}
}
