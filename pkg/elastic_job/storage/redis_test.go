package storage

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedisStorage_Lock(t *testing.T) {
	t.Run("LockTesting", func(t *testing.T) {
		redisStorage, err := NewRedisStorage(&Config{
			Endpoints:   []string{"0.0.0.0:6379"},
			DialTimeout: time.Second,
			Password:    "redis",
		})
		assert.NoError(t, err)

		err = redisStorage.TryLock("k1")
		assert.NoError(t, err)

		// got lock for k1
		// doing...
		done := make(chan struct{})

		go func() {
			redisStorage, err := NewRedisStorage(&Config{
				Endpoints:   []string{"0.0.0.0:6379"},
				DialTimeout: time.Second,
				Password:    "redis",
			})
			assert.NoError(t, err)
			defer func() {
				err = redisStorage.Close()
				assert.NoError(t, err)
			}()

			err = redisStorage.TryLock("k1")
			assert.ErrorIs(t, err, ErrLocked)

			time.Sleep(time.Second * 10)

			err = redisStorage.TryLock("k1")
			assert.NoError(t, err)
			err = redisStorage.UnLock("k1")
			assert.NoError(t, err)
			done <- struct{}{}
		}()

		time.Sleep(time.Second * 5)
		err = redisStorage.UnLock("k1")
		assert.NoError(t, err)

		// wait goroutine
		<-done
		err = redisStorage.Close()
		assert.NoError(t, err)
	})

	t.Run("LockMulti", func(t *testing.T) {
		redisStorage, err := NewRedisStorage(&Config{
			Endpoints:   []string{"0.0.0.0:6379"},
			DialTimeout: time.Second,
			Password:    "redis",
		})
		assert.NoError(t, err)

		err = redisStorage.TryLock("k1")
		assert.NoError(t, err)

		// got lock for k1
		// doing...
		done := make(chan struct{})

		go func() {
			redisStorage, err := NewRedisStorage(&Config{
				Endpoints:   []string{"0.0.0.0:6379"},
				DialTimeout: time.Second,
				Password:    "redis",
			})
			assert.NoError(t, err)
			defer func() {
				err = redisStorage.Close()
				assert.NoError(t, err)
			}()

			err = redisStorage.TryLock("k2")
			assert.NoError(t, err)
			time.Sleep(time.Second)

			err = redisStorage.UnLock("k2")
			assert.NoError(t, err)
			done <- struct{}{}
		}()

		time.Sleep(time.Second)
		err = redisStorage.UnLock("k1")
		assert.NoError(t, err)

		// wait goroutine
		<-done
		err = redisStorage.Close()
		assert.NoError(t, err)
	})
}

func TestRedisStorage_SaveAndWatch(t *testing.T) {
	t.Run("WithNormalStep", func(t *testing.T) {
		redisStorage, err := NewRedisStorage(&Config{
			Endpoints:   []string{"0.0.0.0:6379"},
			DialTimeout: time.Second,
			Password:    "redis",
		})
		assert.NoError(t, err)
		err = redisStorage.Save("k", "v", time.Second)
		assert.NoError(t, err)

		wc := redisStorage.Watch()
		for w := range wc {
			assert.Equal(t, WatchResponse{
				Key:     "k",
				Value:   "v",
				TimeNow: w.TimeNow,
			}, w)
			_ = redisStorage.Close()
			return
		}
	})

	t.Run("WithKeyExpiredAlready", func(t *testing.T) {
		redisStorage, err := NewRedisStorage(&Config{
			Endpoints:   []string{"0.0.0.0:6379"},
			DialTimeout: time.Second,
			Password:    "redis",
		})
		assert.NoError(t, err)
		err = redisStorage.Save("k", "v", time.Second)
		assert.NoError(t, err)

		time.Sleep(time.Second * 2)
		wc := redisStorage.Watch()
		for w := range wc {
			assert.Equal(t, WatchResponse{
				Key:     "k",
				Value:   "v",
				TimeNow: w.TimeNow,
			}, w)
			_ = redisStorage.Close()
			return
		}
	})

	t.Run("WithKeyMulti", func(t *testing.T) {
		redisStorage, err := NewRedisStorage(&Config{
			Endpoints:   []string{"0.0.0.0:6379"},
			DialTimeout: time.Second,
			Password:    "redis",
		})
		assert.NoError(t, err)

		err = redisStorage.Save("k1", "v1", time.Second)
		err = redisStorage.Save("k2", "v2", time.Second*2)
		err = redisStorage.Save("k3", "v3", time.Second*3)
		assert.NoError(t, err)

		wc := redisStorage.Watch()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		i := 1
		for {
			select {
			case w := <-wc:
				assert.Equal(t, WatchResponse{
					Key:     "k" + strconv.Itoa(i),
					Value:   "v" + strconv.Itoa(i),
					TimeNow: w.TimeNow,
				}, w)
				i++
			case <-ctx.Done():
				return
			}
		}
	})
}
