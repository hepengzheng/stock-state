package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
	"github.com/hepengzheng/stock-state/pkg/config"
	"github.com/hepengzheng/stock-state/pkg/logger"
)

var providerSet = wire.NewSet(NewEtcdClient)

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

	ctx, cancel := context.WithCancel(context.Background())
	stockServer, err := initApp(ctx, cfg)
	if err != nil {
		cancel()
		logger.Fatal("init app err", zap.Error(err))
	}

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
