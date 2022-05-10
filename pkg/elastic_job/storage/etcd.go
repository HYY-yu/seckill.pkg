package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	KeyForWatchReversion = "MultiCron/KeyForWatchReversion"
)

type etcdStorage struct {
	ctx    context.Context
	cancel context.CancelFunc

	cfg *Config

	etcdClient  *clientv3.Client
	etcdSession *concurrency.Session

	lockMap map[string]*concurrency.Mutex // key is storage key

	watchC chan WatchResponse

	reversion int64
}

func NewEtcdStorage(config *Config) (BackendStorage, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:            config.Endpoints,
		DialTimeout:          config.DialTimeout,
		Username:             config.Username,
		Password:             config.Password,
		DialKeepAliveTime:    time.Second,
		DialKeepAliveTimeout: 500 * time.Millisecond,
	})
	if err != nil {
		return nil, err
	}
	s1, err := concurrency.NewSession(c)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	b := &etcdStorage{
		ctx:         ctx,
		cancel:      cancel,
		etcdClient:  c,
		etcdSession: s1,
		cfg:         config,
		lockMap:     make(map[string]*concurrency.Mutex),
		watchC:      make(chan WatchResponse),
	}

	go b.run()
	return b, nil
}

func (e *etcdStorage) run() {
	var reversion int64

	getResp, _ := e.etcdClient.Get(e.ctx, KeyForWatchReversion)
	if getResp != nil && getResp.Count > 0 {
		// 有这个key，则读取reversion
		reversion = cast.ToInt64(string(getResp.Kvs[0].Value))
		_, _ = e.etcdClient.KeepAlive(e.ctx, clientv3.LeaseID(getResp.Kvs[0].Lease))
	} else {
		// 没有这个Key ,创建一个新Key
		leaseResp, _ := e.etcdClient.Grant(e.ctx, int64(time.Minute.Seconds()))
		if leaseResp == nil {
			return
		}
		_, _ = e.etcdClient.Put(e.ctx, KeyForWatchReversion, "0", clientv3.WithLease(leaseResp.ID))
	}
	e.reversion = reversion

	watchOpt := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithFilterPut(),
		clientv3.WithPrevKV(),
	}
	if reversion > 0 {
		watchOpt = append(watchOpt, clientv3.WithRev(reversion+1))
	}
	watchChan := e.etcdClient.Watch(e.ctx, KeyPrefixForStorage, watchOpt...)
	for {
		// reversion 持久化
		_, _ = e.etcdClient.Put(e.ctx, KeyForWatchReversion, strconv.Itoa(int(e.reversion)))
		select {
		case wresp := <-watchChan:
			for _, ev := range wresp.Events {
				if ev.Type == clientv3.EventTypeDelete {
					// Key 过期删除事件
					key := string(ev.Kv.Key)
					key = strings.TrimPrefix(key, KeyPrefixForStorage)

					// ETCD 删除事件不会返回Value，需要从PrevKV中取数据
					value := ev.PrevKv.Value
					now := time.Now().Unix()

					select {
					case e.watchC <- WatchResponse{
						Key:     key,
						Value:   string(value),
						TimeNow: now,
					}:
						e.reversion = wresp.Header.Revision // 暂存
					case <-e.ctx.Done():
						return
					}
				}
			}
		case <-e.ctx.Done():
			return
		}
	}
}

func (e *etcdStorage) Save(key string, value string, delay time.Duration) error {
	key = KeyPrefixForStorage + key

	ctx, cancel := context.WithTimeout(e.ctx, e.cfg.DialTimeout)
	defer cancel()
	leaseResp, err := e.etcdClient.Grant(ctx, int64(delay.Seconds()))
	if err != nil {
		return err
	}
	ctx2, cancel2 := context.WithTimeout(e.ctx, e.cfg.DialTimeout)
	defer cancel2()
	_, err = e.etcdClient.Put(ctx2, key, value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	return nil
}

func (e *etcdStorage) Watch() WatchChan {
	return e.watchC
}

func (e *etcdStorage) TryLock(key string) error {
	if _, ok := e.lockMap[key]; ok {
		return ErrLocked
	}
	m1 := concurrency.NewMutex(e.etcdSession, key)
	if err := m1.TryLock(context.TODO()); err == nil {
		e.lockMap[key] = m1
		return nil
	} else if err == concurrency.ErrLocked {
		// already locked in another session
		return ErrLocked
	} else {
		return err
	}
}

func (e *etcdStorage) UnLock(key string) error {
	if _, ok := e.lockMap[key]; !ok {
		return nil
	}
	lock := e.lockMap[key]
	delete(e.lockMap, key)

	err := lock.Unlock(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

func (e *etcdStorage) Close() error {
	// reversion 持久化
	_, _ = e.etcdClient.Put(e.ctx, KeyForWatchReversion, strconv.Itoa(int(e.reversion)))

	if e.cancel != nil {
		e.cancel()
	}

	if e.watchC != nil {
		close(e.watchC)
	}

	var errs []error
	if len(e.lockMap) > 0 {
		for _, lock := range e.lockMap {
			err := lock.Unlock(context.TODO())
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if e.etcdClient != nil {
		err := e.etcdClient.Close()
		if err != nil && err != context.Canceled {
			errs = append(errs, err)
		}
	}
	if e.etcdSession != nil {
		err := e.etcdSession.Close()
		if err != nil && err != context.Canceled {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors: %v", errs)
	}
	return nil
}
