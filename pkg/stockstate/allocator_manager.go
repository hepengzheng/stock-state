package stockstate

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/hepengzheng/stock-state/pkg/config"
	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/storage"
)

var ProviderSet = wire.NewSet(NewAllocatorManager)

type AllocatorManager struct {
	ctx      context.Context
	keySpace sync.Map // key -> allocator
	client   *clientv3.Client

	leadership      *election.Leadership
	storage         storage.StockStorage
	allocatorConfig *config.AllocatorConfig

	clientConns sync.Map // server key -> client stream
}

func NewAllocatorManager(ctx context.Context,
	client *clientv3.Client,
	stockStorage storage.StockStorage,
	cfg *config.Config,
) *AllocatorManager {
	return &AllocatorManager{
		ctx:             ctx,
		client:          client,
		storage:         stockStorage,
		allocatorConfig: cfg.AllocatorConfig,
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

	allocator := NewAllocator(am.ctx, am.client, am.storage, key, am.allocatorConfig)
	_ = allocator.Init()
	am.keySpace.Store(key, allocator)
	return allocator
}

func (am *AllocatorManager) Close() error {
	var errs []error
	am.keySpace.Range(func(key, value interface{}) bool {
		allocator := value.(*Allocator)
		if allocator != nil {
			errs = append(errs, allocator.Close())
		}
		return true
	})
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// TODO : add Delete and Reset
