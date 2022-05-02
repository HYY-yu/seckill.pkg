package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
)

// Redis must be turned on notify-keyspace-events
// see: http://doc.redisfans.com/topic/notification.html
type redisStorage struct {
	cfg *Config

	client *redis.Client
	pubsub *redis.PubSub

	lockMap map[string]*redislock.Lock // key is storage key

	watchC chan WatchResponse
}

func NewRedisStorage(config *Config) (BackendStorage, error) {
	c := redis.NewClient(&redis.Options{
		Addr:        config.Endpoints[0],
		DialTimeout: config.DialTimeout,
		Username:    config.Username,
		Password:    config.Password,
	})

	r := &redisStorage{
		cfg:     config,
		client:  c,
		lockMap: make(map[string]*redislock.Lock),
		watchC:  make(chan WatchResponse),
	}
	sub := r.client.Subscribe(context.TODO(), "__keyevent@0__:expired")
	r.pubsub = sub
	go r.run()
	return r, nil
}

func (r redisStorage) run() {
	cMessage := r.pubsub.Channel()
	for cresp := range cMessage {
		if !strings.HasPrefix(cresp.Payload, KeyPrefixForStorage) {
			continue
		}

		// key has expired
		key := strings.TrimPrefix(cresp.Payload, KeyPrefixForStorage)

		// query key's value
		valueCmd := r.client.Get(context.TODO(), key)
		value := valueCmd.Val() // Value 可能为空！！！
		// 如果长时间没收到 Expired 事件，致使KeyValue过期，Value可能丢失
		now := time.Now().Unix()

		r.watchC <- WatchResponse{
			Key:     key,
			Value:   value,
			TimeNow: now,
		}
	}
}

func (r redisStorage) Save(key string, value string, delay time.Duration) error {
	expiredKey := KeyPrefixForStorage + key

	// 过期事件不会推送Value，所以Value还得另外存
	r.client.SetEX(context.TODO(), expiredKey, "1", delay)

	// 只保存过期时间的二倍，因此Value有可能消失
	// Redis 对数据的保障性不足，推荐使用ETCD
	valueDelay := delay * 2
	r.client.SetEX(context.TODO(), key, value, valueDelay)

	return nil
}

func (r redisStorage) Watch() WatchChan {
	return r.watchC
}

func (r redisStorage) TryLock(key string) error {
	if _, ok := r.lockMap[key]; ok {
		return ErrLocked
	}
	lock, err := redislock.New(r.client).Obtain(context.TODO(), key, 60*time.Second, nil)
	if err == nil {
		r.lockMap[key] = lock
		return nil
	} else if err == redislock.ErrNotObtained {
		return ErrLocked
	} else {
		return err
	}
}

func (r redisStorage) UnLock(key string) error {
	if _, ok := r.lockMap[key]; !ok {
		return nil
	}
	lock := r.lockMap[key]
	delete(r.lockMap, key)

	err := lock.Release(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

func (r redisStorage) Close() error {
	var errs []error

	if len(r.lockMap) > 0 {
		for _, lock := range r.lockMap {
			err := lock.Release(context.TODO())
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if r.pubsub != nil {
		err := r.pubsub.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors: %v", errs)
	}
	return nil
}
