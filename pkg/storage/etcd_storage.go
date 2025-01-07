package storage

import (
	"context"
	"errors"

	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	"github.com/hepengzheng/stock-state/pkg/logger"
)

var ProviderSet = wire.NewSet(NewEtcdStockStorage)

var (
	ErrValueNotFound = errors.New("value not found")
)

type EtcdStockStorage struct {
	client *clientv3.Client
}

func NewEtcdStockStorage(client *clientv3.Client) StockStorage {
	return &EtcdStockStorage{client: client}
}

// Delete implements StockStorage.
func (e *EtcdStockStorage) Delete(ctx context.Context, key string) error {
	_, err := e.client.Delete(ctx, key)
	return err
}

// Load implements StockStorage.
func (e *EtcdStockStorage) Load(ctx context.Context, key string) (int32, error) {
	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if len(resp.Kvs) == 0 {
		return 0, ErrValueNotFound
	}
	return decodeValue(string(resp.Kvs[0].Value))
}

// Save implements StockStorage.
func (e *EtcdStockStorage) Save(ctx context.Context, key string, value int32) error {
	encoded := encodeValue(value)
	resp, err := e.client.Txn(ctx).If(clientv3.Compare(clientv3.Value(key), "<", encoded)).
		Then(clientv3.OpPut(key, encoded)).
		Else(clientv3.OpGet(key)).
		Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		// the If statement returns false
		var storedValue string
		if len(resp.Responses) > 0 {
			kvs := resp.Responses[0].GetResponseRange().Kvs
			if len(kvs) > 0 {
				storedValue = string(kvs[0].Value)
			}
		}
		logger.Error("key may by updated concurrently", zap.String("key", key),
			zap.String("stored_value", storedValue))
	}
	return nil
}
