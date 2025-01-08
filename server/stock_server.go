package server

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/jsonutil"
	"github.com/hepengzheng/stock-state/pkg/logger"
	"github.com/hepengzheng/stock-state/pkg/stockstate"
)

type StockServer struct {
	statepb.UnimplementedStateServer

	ctx context.Context
	am  *stockstate.AllocatorManager

	clientConns sync.Map // server key -> client stream
}

func NewServer(ctx context.Context, am *stockstate.AllocatorManager) *StockServer {
	return &StockServer{
		ctx: ctx,
		am:  am,
	}
}

func (s *StockServer) GetStock(ctx context.Context, req *statepb.Request) (*statepb.Response, error) {
	leadership := s.am.GetLeadership(req.Key)
	if !leadership.IsLeader() {
		logger.Info("try to forward request", zap.String("req", req.String()))
		return s.forwardRequestToLeader(ctx, leadership,
			func(ctx context.Context, client statepb.StateClient) (*statepb.Response, error) {
				return client.GetStock(ctx, req)
			},
		)
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

	logger.Info("get stock finished", zap.Any("resp", jsonutil.MustMarshal(resp)),
		zap.Int32("stock", resp.Result))
	return resp, nil
}

func (s *StockServer) Close() error {
	return s.am.Close()
}

type call func(ctx context.Context, client statepb.StateClient) (*statepb.Response, error)

func (s *StockServer) forwardRequestToLeader(ctx context.Context, leadership *election.Leadership, fn call) (*statepb.Response, error) {
	clientConn, err := s.getDelegateClient(leadership)
	if err != nil {
		logger.Error("failed to get delegate client", zap.Error(err))
		return nil, err
	}

	client := statepb.NewStateClient(clientConn)
	return fn(ctx, client)
}

func (s *StockServer) getDelegateClient(leadership *election.Leadership) (*grpc.ClientConn, error) {
	if leadership == nil {
		return nil, errors.New("failed to get leadership")
	}
	var leaderID string
	for range 3 {
		leaderID = leadership.GetLeaderID()
		if leaderID == "" {
			<-time.After(time.Millisecond * 10)
			continue
		}
	}

	if leaderID == "" {
		logger.Error("failed to get leadership", zap.String("leadership", leadership.GetLeaderID()))
		return nil, errors.New("failed to get leadership")
	}

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
