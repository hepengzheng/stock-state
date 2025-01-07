package storage

import (
	"context"
	"fmt"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestEtcdStockStorage_Save(t *testing.T) {
	key := "test-save"
	type args struct {
		ctx   context.Context
		key   string
		value int32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "first save",
			args: args{
				ctx:   context.Background(),
				key:   key,
				value: 1,
			},
			wantErr: false,
		},
		{
			name: "second save",
			args: args{
				ctx:   context.Background(),
				key:   key,
				value: 2,
			},
			wantErr: false,
		},
		{
			name: "third save, but value is less",
			args: args{
				ctx:   context.Background(),
				key:   key,
				value: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
			if err != nil {
				t.Fatal(err)
			}
			e := &EtcdStockStorage{
				client: client,
			}
			err = e.Save(tt.args.ctx, tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("EtcdStorage.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Cleanup(func() {
		client, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
		if err != nil {
			t.Fatal(err)
		}

		_, _ = client.Delete(context.Background(), key)
	})
}

func TestEtcdStockStorage_Get(t *testing.T) {
	key := "benchmark-key:map"
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		t.Fatal(err)
	}
	e := &EtcdStockStorage{
		client: client,
	}
	res, err := e.Load(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}
