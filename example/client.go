package main

import (
	"context"
	"strconv"
	"sync"
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
		sent     int32 = 0
		wg       sync.WaitGroup
		respChan = make(chan int32)
	)

	go func() {
		defer logutil.LogPanic()
		defer close(respChan)
		for range 8000 {
			wg.Add(1)
			if sent%200 == 0 {
				<-time.After(10 * time.Millisecond)
			}
			sent++
			go func() {
				defer logutil.LogPanic()
				defer wg.Done()
				now := time.Now().UnixMicro()
				req := &statepb.Request{
					RequestId: strconv.FormatInt(now, 10),
					Timestamp: now,
					Key:       "benchmark-key",
					Count:     1,
				}
				resp, err := sendOneRequest(client, req)
				if err != nil {
					logger.Error("Failed to send request", zap.Error(err))
				} else {
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

	logger.Info("finished", zap.Int32("res", res), zap.Int32("sent", sent),
		zap.Int64("elapsed ms", time.Since(start).Milliseconds()))
}

func sendOneRequest(client statepb.StateClient, req *statepb.Request) (*statepb.Response, error) {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()
	return client.GetStock(ctx, req)
}
