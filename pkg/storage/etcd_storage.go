package storage

import (
	"context"
	"errors"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	_, err := e.client.Delete(ctx, key)
	return err
}

// Load implements StockStorage.
func (e *EtcdStockStorage) Load(ctx context.Context, key string) (int32, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if len(resp.Kvs) == 0 {
		return 0, ErrValueNotFound
	}
	return decodeValue(string(resp.Kvs[0].Value)), nil
}

// Save implements StockStorage.
func (e *EtcdStockStorage) Save(ctx context.Context, key string, value int32) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	txn := e.client.Txn(ctx)

	got, err := e.client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get key from etcd, key:%s, err:%w", key, err)
	}
	var currentValue string
	switch kvLen := len(got.Kvs); {
	case kvLen == 0:
		txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0))
	case kvLen == 1:
		currentValue = string(got.Kvs[0].Value)
		txn.If(clientv3.Compare(clientv3.Value(key), "=", currentValue))
	default:
		return fmt.Errorf("incorrect kv len of key:%s", key)
	}

	if currentValue != "" {
		cu := decodeValue(currentValue)
		if cu >= value {
			return fmt.Errorf("cannot save value less or equal to the preview one, key:%s", key)
		}
	}

	resp, err := txn.Then(clientv3.OpPut(key, encodeValue(value))).Commit()
	if err != nil {
		return fmt.Errorf("failed to save key:%s, err:%w", key, err)
	}
	if !resp.Succeeded {
		return fmt.Errorf("etcd update conflict, key:%s", key)
	}

	return nil
}
