package main

import (
	"context"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
	"github.com/hepengzheng/stock-state/pkg/jsonutil"
	"github.com/hepengzheng/stock-state/pkg/logger"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		logger.Error("Failed to connect", zap.Error(err))
		return
	}
	defer conn.Close()

	client := statepb.NewStateClient(conn)

	now := time.Now().UnixMicro()
	req := &statepb.Request{
		RequestId: strconv.FormatInt(now, 10),
		Timestamp: now,
		Key:       "test-key",
		Count:     1,
	}

	sendOneRequest(err, client, req)

	now = time.Now().UnixMicro()
	req = &statepb.Request{
		RequestId: strconv.FormatInt(now, 10),
		Timestamp: now,
		Key:       "test-key",
		Count:     100,
	}
	sendOneRequest(err, client, req)
}

func sendOneRequest(err error, client statepb.StateClient, req *statepb.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetStock(ctx, req)
	if err != nil {
		logger.Error("Error calling SayHello", zap.Error(err))
		return
	}

	logger.Info("Response from server", zap.String("response", jsonutil.MustMarshal(res)))
}
