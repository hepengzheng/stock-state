package config

import clientv3 "go.etcd.io/etcd/client/v3"

type Config struct {
	AllocatorConfig *AllocatorConfig `json:"allocator_config"`
	EtcdConfig      *clientv3.Config `json:"etcd_config"`
}

type AllocatorConfig struct {
	UpdateInterval int32 `json:"update_interval"`
}
