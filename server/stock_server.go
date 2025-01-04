package server

import (
	"context"
	"errors"
	"io"

	"github.com/google/wire"
	"go.uber.org/zap"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
	"github.com/hepengzheng/stock-state/pkg/logger"
	"github.com/hepengzheng/stock-state/pkg/logutil"
	"github.com/hepengzheng/stock-state/pkg/stockstate"
)

var ProviderSet = wire.NewSet(NewServer)

type StockServer struct {
	statepb.UnimplementedStateServer

	ctx context.Context
	am  *stockstate.AllocatorManager
}

func NewServer(ctx context.Context, am *stockstate.AllocatorManager) *StockServer {
	return &StockServer{
		ctx: ctx,
		am:  am,
	}
}

func (s *StockServer) GetStock(stream statepb.State_GetStockServer) error {
	defer logutil.LogPanic()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// TODO(hpz): forward the request to the leader

		resp := &statepb.Response{
			RequestId: req.RequestId,
			Timestamp: req.Timestamp,
		}
		stock, err := s.am.GetStock(context.Background(), req.Key, req.Count)
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

		if err = stream.Send(resp); err != nil {
			logger.Error("failed to send response", zap.Error(err),
				zap.String("key", req.Key), zap.String("request_id", resp.RequestId))
		}
	}
}
