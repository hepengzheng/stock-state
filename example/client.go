package main

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
	"github.com/hepengzheng/stock-state/pkg/logger"
	"github.com/hepengzheng/stock-state/pkg/logutil"
)

func main() {
	defer logutil.LogPanic()
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect", zap.Error(err))
	}
	defer conn.Close()

	client := statepb.NewStateClient(conn)

	start := time.Now()
	var (
		requestID atomic.Int32
		sent      atomic.Int32
		wg        sync.WaitGroup
		respChan  = make(chan int32)
	)
	go func() {
		defer logutil.LogPanic()
		defer close(respChan)
		const numRequest = 6000
		wg.Add(numRequest)
		for range numRequest {
			if sent.Load()%100 == 0 {
				<-time.After(21 * time.Millisecond)
			}
			sent.Add(1)
			go func() {
				defer logutil.LogPanic()
				defer wg.Done()
				now := time.Now().UnixMicro()
				req := &statepb.Request{
					RequestId: strconv.FormatInt(int64(requestID.Add(1)), 10),
					Timestamp: now,
					Key:       "benchmark-key",
					Count:     1,
				}
				resp, err := sendOneRequest(client, req)
				if err != nil {
					logger.Error("Failed to send request", zap.Error(err))
				} else {
					logger.Info("Sent request",
						zap.String("requestId", resp.RequestId),
						zap.Int32("result", resp.Result))
					respChan <- resp.Result
				}
			}()
		}
		wg.Wait()
	}()

	var res int32
	for v := range respChan {
		if v > res {
			res = v
		}
	}

	logger.Info("finished", zap.Int32("res", res), zap.Int32("sent", sent.Load()),
		zap.Int64("elapsed ms", time.Since(start).Milliseconds()))
}

func sendOneRequest(client statepb.StateClient, req *statepb.Request) (*statepb.Response, error) {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()
	return client.GetStock(ctx, req)
}
