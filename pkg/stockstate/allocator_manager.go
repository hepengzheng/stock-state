package stockstate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/wire"

	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/storage"
)

var ProviderSet = wire.NewSet(NewAllocatorManager)

type AllocatorManager struct {
	ctx      context.Context
	keySpace sync.Map // key -> allocator

	leadership     *election.Leadership
	storage        storage.StockStorage
	updateInterval time.Duration
}

func NewAllocatorManager(ctx context.Context,
	leadership *election.Leadership,
	stockStorage storage.StockStorage,
	updateInterval time.Duration,
) *AllocatorManager {
	return &AllocatorManager{
		ctx:            ctx,
		leadership:     leadership,
		storage:        stockStorage,
		updateInterval: updateInterval,
	}
}

func (am *AllocatorManager) GetStock(ctx context.Context, key string, count int32) (int32, error) {
	at, ok := am.keySpace.Load(key)
	if ok {
		allocator, _ok := at.(*Allocator)
		if !_ok || allocator == nil { // should never happen
			return 0, fmt.Errorf("invalid allocator")
		}
		return allocator.Emit(ctx, count)
	}

	allocator := NewAllocator(ctx, am.leadership, am.storage, key,
		&AllocatorConfig{UpdateInterval: am.updateInterval})
	am.keySpace.Store(key, allocator)
	return allocator.Emit(ctx, count)
}

// TODO : add Delete and Reset
