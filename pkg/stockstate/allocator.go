package stockstate

import (
	"context"
	"time"

	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/logutil"
	"github.com/hepengzheng/stock-state/pkg/storage"
)

type AllocatorConfig struct {
	UpdateInterval time.Duration
}

type Allocator struct {
	ctx    context.Context
	cancel context.CancelFunc

	key string

	leadership *election.Leadership
	state      *State
}

func NewAllocator(ctx context.Context,
	leadership *election.Leadership,
	storage storage.StockStorage,
	prefix string,
	config *AllocatorConfig,
) *Allocator {
	m := Allocator{leadership: leadership}
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.state = &State{
		storage:        storage,
		prefix:         prefix,
		updateInterval: config.UpdateInterval,
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

func (at *Allocator) Close() {
	at.cancel()
	_ = at.state.close()
}

func (at *Allocator) Emit(ctx context.Context, count int32) (int32, error) {
	res, err := at.state.Emit(ctx, at.leadership, count)
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
