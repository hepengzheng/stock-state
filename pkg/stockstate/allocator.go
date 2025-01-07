package stockstate

import (
	"context"
	"fmt"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/hepengzheng/stock-state/pkg/config"
	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/hostutil"
	"github.com/hepengzheng/stock-state/pkg/logutil"
	"github.com/hepengzheng/stock-state/pkg/storage"
)

type Allocator struct {
	ctx    context.Context
	cancel context.CancelFunc

	key string

	leadership *election.Leadership
	state      *State

	clientConns sync.Map
}

func NewAllocator(ctx context.Context,
	client *clientv3.Client,
	storage storage.StockStorage,
	prefix string,
	config *config.AllocatorConfig,
) *Allocator {
	m := Allocator{
		leadership: election.NewLeadership(client,
			fmt.Sprintf("/leader-%s", prefix),
			hostutil.GetLocalAddr(),
		),
	}
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.state = &State{
		storage:        storage,
		prefix:         prefix,
		updateInterval: time.Duration(config.UpdateInterval) * time.Millisecond,
		ctx:            m.ctx,
	}
	return &m
}

func (at *Allocator) Init() error {
	err := at.state.Init(at.ctx)
	if err != nil {
		return err
	}

	go at.UpdateLoop()
	return nil
}

func (at *Allocator) Close() error {
	at.cancel()
	_ = at.state.close()
	return nil
}

func (at *Allocator) GetStock(ctx context.Context, count int32) (int32, error) {
	res, err := at.state.GetStock(ctx, at.leadership, count)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (at *Allocator) Delete(ctx context.Context) error {
	return at.state.Delete(ctx)
}

func (at *Allocator) UpdateLoop() {
	defer logutil.LogPanic()
	ticker := time.NewTicker(at.state.updateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-at.ctx.Done():
			return
		case <-ticker.C:
			if err := at.state.Update(); err != nil {
			}
		}
	}
}

func (at *Allocator) IsLeader() bool {
	return at.leadership.IsLeader()
}
