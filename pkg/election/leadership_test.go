package election

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestLeadership(t *testing.T) {
	cfg := clientv3.Config{Endpoints: []string{"localhost:2379"}}
	cli, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	const leaderKey = "/leader/"
	ld1 := NewLeadership(cli, leaderKey, "server1")
	ld2 := NewLeadership(cli, leaderKey, "server2")

	ctx, cancel := context.WithCancel(context.Background())
	go ld1.Start(ctx)
	go func() {
		<-time.After(time.Second)
		ld2.Start(ctx)
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if ld1.IsLeader() {
					fmt.Printf("[%s]server1 is the leader\n", now())
				}
			case <-ctx.Done():
				return
			}

		}
	}()

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if ld2.IsLeader() {
					fmt.Printf("[%s]server2 is the leader\n", now())
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	<-time.After(10 * time.Second)
	ld1.Stop(context.Background())

	<-time.After(10 * time.Second)
	cancel()

	wg.Wait()

	ld2.Stop(context.Background())
	fmt.Println("end...")
}

func now() string {
	return time.Now().Format(time.DateTime)
}
