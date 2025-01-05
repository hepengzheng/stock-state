package server

import (
	"context"
	"errors"
	"sync"

	"github.com/google/wire"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/logger"
	"github.com/hepengzheng/stock-state/pkg/stockstate"
)

var ProviderSet = wire.NewSet(NewServer)

type StockServer struct {
	statepb.UnimplementedStateServer

	ctx context.Context
	am  *stockstate.AllocatorManager

	leadership  *election.Leadership
	clientConns sync.Map // server key -> client stream
}

func NewServer(ctx context.Context,
	leadership *election.Leadership,
	am *stockstate.AllocatorManager,
) *StockServer {
	return &StockServer{
		ctx:        ctx,
		am:         am,
		leadership: leadership,
	}
}

func (s *StockServer) GetStock(ctx context.Context, req *statepb.Request) (*statepb.Response, error) {
	if !s.leadership.IsLeader() {
		return s.forwardRequestToLeader(ctx, func(ctx context.Context, client statepb.StateClient) (*statepb.Response, error) {
			return client.GetStock(ctx, req)
		})
	}

	resp := &statepb.Response{
		RequestId: req.RequestId,
		Timestamp: req.Timestamp,
	}
	stock, err := s.am.GetStock(ctx, req.Key, req.Count)
	if err != nil {
		switch {
		case errors.Is(err, stockstate.ErrNotInitialized):
			resp.Error = statepb.Response_SERVER_NOT_INITIALIZED
		case errors.Is(err, stockstate.ErrRetryExceeded):
			resp.Error = statepb.Response_SERVER_RETRY_LIMIT_EXCEEDED
		default:
			resp.Error = statepb.Response_UNKNOWN
			logger.Error("failed to get stock", zap.Error(err),
				zap.String("key", req.Key), zap.String("request_id", resp.RequestId))
		}
	} else {
		resp.Error = statepb.Response_OK
		resp.Result = stock
	}

	return resp, nil
}

type call func(ctx context.Context, client statepb.StateClient) (*statepb.Response, error)

func (s *StockServer) forwardRequestToLeader(ctx context.Context, fn call) (*statepb.Response, error) {
	clientConn, err := s.getDelegateClient()
	if err != nil {
		logger.Error("failed to get delegate client", zap.Error(err))
		return nil, err
	}

	client := statepb.NewStateClient(clientConn)
	return fn(ctx, client)
}

func (s *StockServer) getDelegateClient() (*grpc.ClientConn, error) {
	leaderID := s.leadership.GetLeaderID()
	clientConn, ok := s.clientConns.Load(leaderID)
	if ok {
		return clientConn.(*grpc.ClientConn), nil
	}

	newConn, err := grpc.DialContext(s.ctx, leaderID, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	conn, load := s.clientConns.LoadOrStore(leaderID, newConn)
	if !load {
		return newConn, nil
	}

	if err = newConn.Close(); err != nil {
		logger.Warn("failed to close client connection", zap.Error(err))
	}
	return conn.(*grpc.ClientConn), err
}
