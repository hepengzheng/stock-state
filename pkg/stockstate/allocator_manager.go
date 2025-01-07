package stockstate

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	"github.com/hepengzheng/stock-state/pkg/config"
	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/logger"
	"github.com/hepengzheng/stock-state/pkg/storage"
)

var ProviderSet = wire.NewSet(NewAllocatorManager)

type keyManager struct {
	sync.Mutex
	keySpace map[string]*Allocator
}

type AllocatorManager struct {
	ctx    context.Context
	client *clientv3.Client

	keyManager *keyManager

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
		keyManager:      &keyManager{keySpace: make(map[string]*Allocator)},
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
	am.keyManager.Lock()
	defer am.keyManager.Unlock()
	at, ok := am.keyManager.keySpace[key]
	if ok {
		return at
	}

	allocator := NewAllocator(am.ctx, am.client, am.storage, key, am.allocatorConfig)
	err := allocator.Init()
	if err != nil {
		logger.Error("failed to init allocator", zap.String("key", key), zap.Error(err))
	}
	am.keyManager.keySpace[key] = allocator
	return allocator
}

func (am *AllocatorManager) Close() error {
	var errs []error
	am.keyManager.Lock()
	defer am.keyManager.Unlock()
	for _, at := range am.keyManager.keySpace {
		errs = append(errs, at.Close())
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// TODO : add Delete and Reset
