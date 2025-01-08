package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
	"github.com/hepengzheng/stock-state/pkg/config"
	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/hostutil"
	"github.com/hepengzheng/stock-state/pkg/logger"
	"github.com/hepengzheng/stock-state/pkg/stockstate"
	"github.com/hepengzheng/stock-state/pkg/storage"
	"github.com/hepengzheng/stock-state/server"
)

func NewEtcdClient(cfg *config.Config) (*clientv3.Client, error) {
	return clientv3.New(*cfg.EtcdConfig)
}

func main() {
	const port = ":9000" // TODO(hpz): put it into config files

	cfg := &config.Config{
		AllocatorConfig: &config.AllocatorConfig{
			UpdateInterval: 100,
		},
		EtcdConfig: &clientv3.Config{Endpoints: []string{"localhost:2379"}},
	}

	client, err := NewEtcdClient(cfg)
	if err != nil {
		logger.Fatal("failed to create etcd client", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	leadership := election.NewLeadership(client, "leader", hostutil.GetLocalAddr())
	err = leadership.Start(ctx)
	if err != nil {
		logger.Error("failed to start leader", zap.Error(err))
	}
	stockStorage := storage.NewEtcdStockStorage(client)
	allocatorManager := stockstate.NewAllocatorManager(ctx, client, stockStorage, cfg, leadership)
	stockServer := server.NewServer(ctx, allocatorManager)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	grpcServer := grpc.NewServer()
	statepb.RegisterStateServer(grpcServer, stockServer)

	logger.Info("gRPC server is running...", zap.String("port", port))
	if err = grpcServer.Serve(listener); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signalChan
	if err = stockServer.Close(); err != nil {
		logger.Error("close server err", zap.Error(err))
	}
	cancel()

	logger.Info("gRPC server stopped")
}
