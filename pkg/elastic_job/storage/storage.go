package storage

import (
	"errors"
	"time"
)

var (
	ErrLocked = errors.New("already locked. ")
)

const KeyPrefixForStorage = "MultiCron/StoragePrefix"

type WatchResponse struct {
	Key     string
	Value   string
	TimeNow int64
}

type WatchChan <-chan WatchResponse

type BackendStorage interface {
	// Save 存储器，并开始计时
	// TTL 时间后，key删除事件通过 Watch 推送
	Save(key string, value string, delay time.Duration) error
	// Watch 监听删除事件回调
	Watch() WatchChan

	// TryLock  分布式锁
	TryLock(key string) error
	// UnLock   分布式锁
	UnLock(key string) error

	Close() error
}

type Type string

const (
	ETCD  Type = "ETCD"
	Redis Type = "REDIS"
)

type Config struct {
	Endpoints   []string
	DialTimeout time.Duration

	Username string
	Password string
}
