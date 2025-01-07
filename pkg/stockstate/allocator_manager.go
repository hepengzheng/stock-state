package stockstate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/storage"
)

var ProviderSet = wire.NewSet(NewAllocatorManager)

type AllocatorManager struct {
	ctx      context.Context
	keySpace sync.Map // key -> allocator
	client   *clientv3.Client

	leadership     *election.Leadership
	storage        storage.StockStorage
	updateInterval time.Duration

	clientConns sync.Map // server key -> client stream
}

func NewAllocatorManager(ctx context.Context,
	client *clientv3.Client,
	stockStorage storage.StockStorage,
	updateInterval time.Duration,
) *AllocatorManager {
	return &AllocatorManager{
		ctx:            ctx,
		client:         client,
		storage:        stockStorage,
		updateInterval: updateInterval,
	}
}

func (am *AllocatorManager) GetStock(ctx context.Context, key string, count int32) (int32, error) {
	allocator := am.getOrCreateAllocator(key)
	if allocator == nil {
		return 0, fmt.Errorf("invalid allocator for key: %s", key) // should never happen
	}

	return allocator.GetStock(ctx, count)
}

func (am *AllocatorManager) GetLeadership(key string) *election.Leadership {
	allocator := am.getOrCreateAllocator(key)
	if allocator == nil {
		return nil
	}
	return allocator.leadership
}

func (am *AllocatorManager) getOrCreateAllocator(key string) *Allocator {
	at, ok := am.keySpace.Load(key)
	if ok {
		return at.(*Allocator)
	}

	allocator := NewAllocator(am.ctx, am.client, am.storage, key,
		&AllocatorConfig{UpdateInterval: am.updateInterval})
	am.keySpace.Store(key, allocator)
	return allocator
}

// TODO : add Delete and Reset
