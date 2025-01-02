package stockstate

import (
	"context"
	"time"

	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/logutil"
	"github.com/hepengzheng/stock-state/pkg/storage"
)

type ManagerConfig struct {
	UpdateInterval time.Duration
}

type Manger struct {
	ctx    context.Context
	cancel context.CancelFunc

	key string

	leadership *election.Leadership
	state      *State
}

func NewManger(ctx context.Context,
	leadership *election.Leadership,
	storage storage.StockStorage,
	prefix string,
	config *ManagerConfig,
) *Manger {
	m := Manger{leadership: leadership}
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.state = &State{
		storage:        storage,
		prefix:         prefix,
		updateInterval: config.UpdateInterval,
		ctx:            m.ctx,
	}
	return &m
}

func (m *Manger) Init() error {
	err := m.state.Init(m.ctx)
	if err != nil {
		return err
	}

	go m.UpdateLoop()
	return nil
}

func (m *Manger) Close() {
	m.cancel()
	_ = m.state.close()
}

func (m *Manger) Emit(ctx context.Context, count int32) (int32, error) {
	res, err := m.state.Emit(ctx, m.leadership, count)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (m *Manger) Delete(ctx context.Context) error {
	return m.state.Delete(ctx)
}

func (m *Manger) UpdateLoop() {
	defer logutil.LogPanic()
	ticker := time.NewTicker(m.state.updateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if err := m.state.Update(); err != nil {
			}
		}
	}
}
