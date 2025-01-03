package storage

import (
	"context"

	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var ProviderSet = wire.NewSet(NewEtcdStockStorage)

type EtcdStockStorage struct {
	client *clientv3.Client
}

func NewEtcdStockStorage(client *clientv3.Client) StockStorage {
	return &EtcdStockStorage{client: client}
}

// Delete implements StockStorage.
func (e *EtcdStockStorage) Delete(ctx context.Context, key string) error {
	panic("unimplemented")
}

// Load implements StockStorage.
func (e *EtcdStockStorage) Load(ctx context.Context, key string) (int32, error) {
	panic("unimplemented")
}

// Save implements StockStorage.
func (e *EtcdStockStorage) Save(ctx context.Context, key string, value int32) error {
	panic("unimplemented")
}
