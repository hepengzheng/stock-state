package election

import (
	"context"
	"fmt"
	"sync/atomic"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/hepengzheng/stock-state/pkg/logutil"
)

const defaultSessionTTL = 5 // seconds

type Leadership struct {
	client *clientv3.Client

	purpose string

	// leaderKey and leaderID will not change once initilized
	leaderKey string
	leaderID  string // we may use the container IP as leaderID

	session  *concurrency.Session
	election *concurrency.Election

	// currentLeaderID stores the id of the current leader
	currentLeaderID atomic.Value
}

func NewLeadership(client *clientv3.Client, purpose, leaderKey, leaderID string) *Leadership {
	return &Leadership{
		client:    client,
		purpose:   purpose,
		leaderKey: leaderKey,
		leaderID:  leaderID,
	}
}

func (ls *Leadership) Start(ctx context.Context) error {
	defer logutil.LogPanic()
	sess, err := concurrency.NewSession(ls.client, concurrency.WithTTL(defaultSessionTTL))
	if err != nil {
		return err
	}
	ls.session = sess
	ls.election = concurrency.NewElection(sess, ls.leaderKey)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("leadership exits because %v", ctx.Err())
		default:
			// the following line will block until the current server becomes the leader
			err = ls.election.Campaign(ctx, ls.leaderID)
			if err != nil {
				return err
			}
			ls.currentLeaderID.Store(ls.leaderID)

			// this line will block until the ctx is canceled or the watched key is deleted by etcd
			ls.watchLeadership(ctx)
		}
	}
}

func (ls *Leadership) Stop(ctx context.Context) {
	ls.currentLeaderID = atomic.Value{}
	_ = ls.election.Resign(ctx)
	if ls.session != nil {
		err := ls.session.Close()
		if err != nil { // TODO(log)
		}
	}
}

func (ls *Leadership) IsLeader() bool {
	if ls == nil {
		return false
	}
	v := ls.currentLeaderID.Load()
	if v == nil {
		return false
	}
	return v.(string) == ls.leaderID
}

func (ls *Leadership) watchLeadership(ctx context.Context) {
	observer := ls.election.Observe(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-observer:
			if !ok {
				return
			}
			ls.currentLeaderID.Store(string(resp.Kvs[0].Value))
		}
	}
}
