package server

import (
	"context"
	"errors"
	"io"

	"github.com/google/wire"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
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

		resp := &statepb.Response{
			RequestId: req.RequestId,
			Timestamp: req.Timestamp,
		}
		stock, err := s.am.GetStock(context.Background(), req.Key, req.Count)
		if err != nil {
			if errors.Is(err, stockstate.ErrNotInitialized) {
				resp.Error = statepb.Response_SERVER_NOT_INITIALIZED
			} else if errors.Is(err, stockstate.ErrRetryExceeded) {
				resp.Error = statepb.Response_SERVER_RETRY_LIMIT_EXCEEDED
			}
		} else {
			resp.Error = statepb.Response_OK
			resp.Result = stock
		}

		if err = stream.Send(resp); err != nil {
			// TODO(hpz): log error
		}
	}
}
