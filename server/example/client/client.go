package main

import (
	"context"
	"io"
	"log"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	statepb "github.com/hepengzheng/stock-state/api/statebp"
	"github.com/hepengzheng/stock-state/pkg/jsonutil"
	"github.com/hepengzheng/stock-state/pkg/logger"
	"github.com/hepengzheng/stock-state/pkg/logutil"
)

func main() {
	conn, err := grpc.DialContext(context.Background(), "localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := statepb.NewStateClient(conn)
	stream, err := client.GetStock(context.Background())
	if err != nil {
		log.Fatalf("frror creating stream: %v", err)
	}

	done := make(chan struct{})

	go func() {
		defer logutil.LogPanic()
		defer close(done)
		for {
			response, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger.Error("error receiving response", zap.Error(err))
				break
			}
			log.Printf("server: %s\n", jsonutil.MustMarshal(response))
		}
	}()

	count := 0
	for {
		now := time.Now().UnixMilli()
		err = stream.Send(&statepb.Request{
			RequestId: strconv.FormatInt(now, 10),
			Timestamp: now,
			Key:       "test-key",
			Count:     1,
		})
		if err != nil {
			logger.Error("error sending message", zap.Error(err))
			break
		}

		count++

		if count >= 10 {
			break
		}
	}

	err = stream.CloseSend()
	if err != nil {
		logger.Error("error closing stream", zap.Error(err))
	}
	<-done
}
