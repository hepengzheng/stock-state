package stockstate

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hepengzheng/stock-state/pkg/election"
	"github.com/hepengzheng/stock-state/pkg/storage"
)

const (
	defaultAllocCount = 500

	capValueKey = "cap"
)

var (
	ErrNotInitialized = errors.New("state is not initialized")
	ErrNotLeader      = errors.New("server is not the leader")
	ErrCapReached     = errors.New("cap reached")
	ErrRetryExceeded  = errors.New("retry exceeded")
)

type State struct {
	sync.Mutex
	storage storage.StockStorage

	ctx            context.Context
	prefix         string
	updateInterval time.Duration

	value    atomic.Int32
	capValue atomic.Int32

	initialized atomic.Bool
}

func (s *State) Init(ctx context.Context) error {
	s.Lock()
	defer s.Unlock()

	if s.initialized.Load() {
		return nil
	}

	// recover from etcd
	capKey := s.getCapKey()
	currentCapValue, err := s.storage.Load(ctx, capKey)
	if err != nil {
		if !errors.Is(err, storage.ErrValueNotFound) {
			return err
		}
		currentCapValue = 0
	}
	newCapValue := currentCapValue + defaultAllocCount
	err = s.storage.Save(ctx, capKey, newCapValue)
	if err != nil {
		return err
	}
	s.value.Store(currentCapValue)
	s.capValue.Store(newCapValue)

	s.initialized.Store(true)
	return nil
}

func (s *State) GetStock(ctx context.Context, leadership *election.Leadership, count int32) (int32, error) {
	const retryLimit = 10
	for range retryLimit {
		select {
		case <-s.ctx.Done():
			return 0, s.ctx.Err()
		case <-ctx.Done():
			return 0, s.ctx.Err()
		default:
		}

		if !s.initialized.Load() {
			if leadership.IsLeader() {
				<-time.After(10 * time.Millisecond)
				continue
			}
			return 0, ErrNotLeader
		}

		res, err := s.getStock(count)
		if err != nil {
			if errors.Is(err, ErrCapReached) {
				<-time.After(s.updateInterval)
				continue
			}
			return 0, err
		}

		if !leadership.IsLeader() {
			return 0, ErrNotLeader
		}
		return res, nil
	}
	return 0, ErrRetryExceeded
}

func (s *State) Update() error {
	if !s.initialized.Load() {
		return ErrNotInitialized
	}

	capKey := s.getCapKey()
	capValue, err := s.storage.Load(s.ctx, capKey)
	if err != nil {
		return err
	}

	lag := capValue - s.value.Load()
	// still serve with the memory counter
	if lag > defaultAllocCount/3 {
		return nil
	}

	// renew the cap
	newCapValue := capValue + defaultAllocCount
	err = s.storage.Save(s.ctx, capKey, newCapValue)
	if err != nil {
		return err
	}
	s.capValue.Store(newCapValue)
	return nil
}

func (s *State) Delete(ctx context.Context) error {
	return s.storage.Delete(ctx, s.prefix)
}

func (s *State) close() error {
	s.Lock()
	defer s.Unlock()
	s.initialized.Store(false)
	return nil
}

func (s *State) getStock(count int32) (int32, error) {
	s.Lock()
	defer s.Unlock()
	value := s.value.Load()
	capValue := s.capValue.Load()
	if capValue-value > count {
		s.value.Add(count)
		return value, nil
	}
	return 0, ErrCapReached
}

func (s *State) getCapKey() string {
	return fmt.Sprintf("%s:%s", s.prefix, capValueKey)
}
