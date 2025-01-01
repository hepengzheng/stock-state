package stockstate

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hepengzheng/stock-state/pkg/logutil"
	"github.com/hepengzheng/stock-state/pkg/storage"
)

const defaultFirstAlloc = 500

var (
	ErrNotInitilized = errors.New("state is not initilized")
)

type State struct {
	sync.Mutex
	storage        storage.StockStorage
	key            string
	value          atomic.Int32
	lastSavedValue atomic.Int32
	initilized     atomic.Bool

	ctx context.Context
}

func NewState(ctx context.Context, storage storage.StockStorage, key string) *State {
	return &State{
		storage: storage,
		key:     key,
		ctx:     ctx,
	}
}

func (state *State) Init(ctx context.Context) error {
	state.Lock()
	defer state.Unlock()

	if state.initilized.Load() {
		return nil
	}

	value, err := state.storage.Load(ctx, state.key)
	if err != nil {
		return err
	}
	newValue := value + defaultFirstAlloc
	err = state.storage.Save(ctx, state.key, newValue)
	if err != nil {
		return err
	}
	state.value.Store(value)
	state.lastSavedValue.Store(value)
	state.initilized.Store(true)
	return nil
}

func (state *State) Get(ctx context.Context, count int32) (int32, error) {
	state.Lock()
	defer state.Unlock()

	if !state.initilized.Load() {
		return 0, ErrNotInitilized
	}

	value := state.value.Load()
	lastSaved := state.lastSavedValue.Load()
	if lastSaved-value > count {
		state.value.Add(count)
		return value, nil
	}

	newValue := value + count + defaultFirstAlloc
	err := state.storage.Save(ctx, state.key, newValue)
	if err != nil {
		return 0, err
	}

	state.value.Add(count)
	state.lastSavedValue.Store(newValue)
	return value + count, nil
}

func (state *State) Update() {
	defer logutil.LogPanic()

	ticker := time.NewTicker(30 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-state.ctx.Done():
			return
		case <-ticker.C:
			// TODO: renew the value
		}
	}
}

// Sync flushes the state to etcd
func (state *State) Sync() error {
	panic("implement me")
}

func (state *State) Close() error {
	panic("implement me")
}
