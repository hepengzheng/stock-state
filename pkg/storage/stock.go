package storage

import "context"

type StockStorage interface {
	Load(ctx context.Context, key string) (int32, error)
	Save(ctx context.Context, key string, value int32) error
	Delete(ctx context.Context, key string) error
}

var _ StockStorage = (*EtcdStockStorage)(nil)
