package cache

// cache 使用 go-redis 提供的 Trace
import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
)

type Repo interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, ttl time.Duration) bool
	ExpireAt(ctx context.Context, key string, ttl time.Time) bool
	Del(ctx context.Context, key string) bool
	Exists(ctx context.Context, keys ...string) bool
	Incr(ctx context.Context, key string) int64
	Client() *redis.Client
	Close() error
}

type cacheRepo struct {
	serverName string
	client     *redis.Client
}

type RedisConf struct {
	Addr         string
	Pass         string
	Db           int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
}

func New(serverName string, cfg *RedisConf) (Repo, error) {
	client, err := redisConnect(serverName, cfg)
	if err != nil {
		return nil, err
	}

	return &cacheRepo{
		serverName: serverName,
		client:     client,
	}, nil
}

func redisConnect(serverName string, cfg *RedisConf) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Pass,
		DB:           cfg.Db,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})
	client.AddHook(redisotel.NewTracingHook(redisotel.WithAttributes(
		attribute.String("servername", serverName),
	)))
	collect := NewPoolStatsCollector(client, serverName)
	_ = prometheus.Register(collect)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis error %w", err)
	}
	return client, nil
}

// Set set some <key,value> into redis
func (c *cacheRepo) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	var err error
	if err = c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		err = fmt.Errorf("redis set key: %s err: %w ", key, err)
	}
	return err
}

// Get run the get command from redis
func (c *cacheRepo) Get(ctx context.Context, key string) (string, error) {
	var err error
	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		err = fmt.Errorf("redis get key: %s err %w", key, err)
	}
	return value, err
}

// TTL get some key from redis
func (c *cacheRepo) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return -1, fmt.Errorf("redis get key: %s err %w", key, err)
	}

	return ttl, nil
}

// Expire expire some key
func (c *cacheRepo) Expire(ctx context.Context, key string, ttl time.Duration) bool {
	ok, _ := c.client.Expire(ctx, key, ttl).Result()
	return ok
}

// ExpireAt expire some key at some time
func (c *cacheRepo) ExpireAt(ctx context.Context, key string, ttl time.Time) bool {
	ok, _ := c.client.ExpireAt(ctx, key, ttl).Result()
	return ok
}

func (c *cacheRepo) Exists(ctx context.Context, keys ...string) bool {
	if len(keys) == 0 {
		return true
	}
	value, _ := c.client.Exists(ctx, keys...).Result()
	return value > 0
}

func (c *cacheRepo) Del(ctx context.Context, key string) bool {
	var err error
	if key == "" {
		return true
	}
	value, err := c.client.Del(ctx, key).Result()
	if err != nil {
		err = fmt.Errorf("redis del key: %s err %w ", key, err)
	}
	return value > 0
}

func (c *cacheRepo) Incr(ctx context.Context, key string) int64 {
	var err error
	value, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		err = fmt.Errorf("redis Incr key: %s err %w ", key, err)
	}
	return value
}

func (c *cacheRepo) Client() *redis.Client {
	return c.client
}

// Close close redis client
func (c *cacheRepo) Close() error {
	return c.client.Close()
}
