package election

import (
	"context"
	"sync/atomic"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"

	"github.com/hepengzheng/stock-state/pkg/logger"
	"github.com/hepengzheng/stock-state/pkg/logutil"
)

const defaultSessionTTL = 3 // seconds

type Leadership struct {
	client *clientv3.Client

	// leaderKey and leaderID will not change once initilized
	leaderKey string
	leaderID  string // we may use the container IP as leaderID

	session  *concurrency.Session
	election *concurrency.Election

	// currentLeaderID stores the id of the current leader
	currentLeaderID atomic.Value
}

func NewLeadership(client *clientv3.Client, leaderKey, leaderID string) *Leadership {
	return &Leadership{
		client:    client,
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
	leaderResp, _ := ls.election.Leader(ctx)
	if leaderResp != nil && len(leaderResp.Kvs) > 0 {
		if cl := string(leaderResp.Kvs[0].Value); cl != "" {
			logger.Info("starting leadership", zap.String("current_leader", cl))
			ls.currentLeaderID.Store(cl)
		}
	}
	go func() {
		defer logutil.LogPanic()
		for {
			select {
			case <-ctx.Done():
				logger.Info("stopping leadership", zap.String("leader_key", ls.leaderKey))
				return
			default:
				// the following line will block until the current server becomes the leader
				err = ls.election.Campaign(ctx, ls.leaderID)
				if err != nil {
					logger.Error("election campaign failed", zap.Error(err))
					return
				}
				ls.currentLeaderID.Store(ls.leaderID)

				// this line will block until the ctx is canceled or the watched key is deleted by etcd
				ls.watchLeadership(ctx)
			}
		}
	}()
	return nil
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

func (ls *Leadership) GetLeaderID() string {
	leaderID := ls.currentLeaderID.Load()
	if leaderID == nil {
		return ""
	}
	return leaderID.(string)
}

func (ls *Leadership) IsLeader() bool {
	if ls == nil {
		return false
	}
	v := ls.currentLeaderID.Load()
	if v == nil {
		return false
	}
	vs := v.(string)
	if vs == ls.leaderID {
		return true
	}
	logger.Info("current leader", zap.String("leaderID", vs))
	return false
}

func (ls *Leadership) watchLeadership(ctx context.Context) {
	observer := ls.election.Observe(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-observer:
			if !ok {
				logger.Info("leader election closed")
				return
			}
			leaderID := string(resp.Kvs[0].Value)
			logger.Info("observed leadership changes", zap.String("leader_id", leaderID))
			ls.currentLeaderID.Store(leaderID)
		}
	}
}
