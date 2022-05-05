package storage

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEtcdStorage_Lock(t *testing.T) {
	t.Run("LockTesting", func(t *testing.T) {
		etcdStorage, err := NewEtcdStorage(&Config{
			Endpoints:   []string{"0.0.0.0:2379", "0.0.0.0:12379", "0.0.0.0:22379"},
			DialTimeout: time.Second,
		})
		assert.NoError(t, err)

		err = etcdStorage.TryLock("k1")
		assert.NoError(t, err)

		// got lock for k1
		// doing...
		done := make(chan struct{})

		go func() {
			etcdStorage, err := NewEtcdStorage(&Config{
				Endpoints:   []string{"0.0.0.0:2379", "0.0.0.0:12379", "0.0.0.0:22379"},
				DialTimeout: time.Second,
			})
			assert.NoError(t, err)
			defer func() {
				err = etcdStorage.Close()
				assert.NoError(t, err)
			}()

			err = etcdStorage.TryLock("k1")
			assert.ErrorIs(t, err, ErrLocked)

			time.Sleep(time.Second * 2)

			err = etcdStorage.TryLock("k1")
			assert.NoError(t, err)
			err = etcdStorage.UnLock("k1")
			assert.NoError(t, err)
			done <- struct{}{}
		}()

		time.Sleep(time.Second)
		err = etcdStorage.UnLock("k1")
		assert.NoError(t, err)

		// wait goroutine
		<-done
		err = etcdStorage.Close()
		assert.NoError(t, err)
	})

	t.Run("LockMulti", func(t *testing.T) {
		etcdStorage, err := NewEtcdStorage(&Config{
			Endpoints:   []string{"0.0.0.0:2379", "0.0.0.0:12379", "0.0.0.0:22379"},
			DialTimeout: time.Second,
		})
		assert.NoError(t, err)

		err = etcdStorage.TryLock("k1")
		assert.NoError(t, err)

		// got lock for k1
		// doing...
		done := make(chan struct{})

		go func() {
			etcdStorage, err := NewEtcdStorage(&Config{
				Endpoints:   []string{"0.0.0.0:2379", "0.0.0.0:12379", "0.0.0.0:22379"},
				DialTimeout: time.Second,
			})
			assert.NoError(t, err)
			defer func() {
				err = etcdStorage.Close()
				assert.NoError(t, err)
			}()

			err = etcdStorage.TryLock("k2")
			assert.NoError(t, err)
			time.Sleep(time.Second)

			err = etcdStorage.UnLock("k2")
			assert.NoError(t, err)
			done <- struct{}{}
		}()

		time.Sleep(time.Second)
		err = etcdStorage.UnLock("k1")
		assert.NoError(t, err)

		// wait goroutine
		<-done
		err = etcdStorage.Close()
		assert.NoError(t, err)
	})
}

func TestEtcdStorage_SaveAndWatch(t *testing.T) {
	t.Run("WithNormalStep", func(t *testing.T) {
		etcdStorage, err := NewEtcdStorage(&Config{
			Endpoints:   []string{"0.0.0.0:2379", "0.0.0.0:12379", "0.0.0.0:22379"},
			DialTimeout: time.Second,
		})
		assert.NoError(t, err)
		defer func() {
			err = etcdStorage.Close()
			assert.NoError(t, err)
		}()

		err = etcdStorage.Save("k", "v", time.Second)
		assert.NoError(t, err)

		wc := etcdStorage.Watch()
		for w := range wc {
			assert.Equal(t, WatchResponse{
				Key:     "k",
				Value:   "v",
				TimeNow: w.TimeNow,
			}, w)
			return
		}
	})

	t.Run("WithKeyExpiredAlready", func(t *testing.T) {
		etcdStorage, err := NewEtcdStorage(&Config{
			Endpoints:   []string{"0.0.0.0:2379", "0.0.0.0:12379", "0.0.0.0:22379"},
			DialTimeout: time.Second,
		})
		assert.NoError(t, err)
		err = etcdStorage.Save("key", "value", time.Second*1)
		err = etcdStorage.Save("key2", "value2", time.Second*3)
		assert.NoError(t, err)

		time.Sleep(time.Second * 2)
		_ = <-etcdStorage.Watch()

		// 意外关闭
		err = etcdStorage.Close()
		assert.NoError(t, err)

		time.Sleep(time.Second * 3) // Key2 过期，但是已经没有人监听 Key2 过期事件

		// 模拟机器重启
		t.Log("Restart Watch... ")
		etcdStorage2, err := NewEtcdStorage(&Config{
			Endpoints:   []string{"0.0.0.0:2379", "0.0.0.0:12379", "0.0.0.0:22379"},
			DialTimeout: time.Second,
		})
		assert.NoError(t, err)
		defer func() {
			err = etcdStorage2.Close()
			assert.NoError(t, err)
		}()

		wc := etcdStorage2.Watch()
		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		for {
			select {
			case w := <-wc:
				assert.Equal(t, WatchResponse{
					Key:     "key2",
					Value:   "value2",
					TimeNow: w.TimeNow,
				}, w)
				// 成功读到之前的key
				return
			case <-ctx.Done():
				return
			}
		}
	})

	t.Run("WithKeyMulti", func(t *testing.T) {
		etcdStorage, err := NewEtcdStorage(&Config{
			Endpoints:   []string{"0.0.0.0:2379", "0.0.0.0:12379", "0.0.0.0:22379"},
			DialTimeout: time.Second,
		})
		assert.NoError(t, err)

		err = etcdStorage.Save("k1", "v1", time.Second)

		// 如果K1和K2过期时间靠的太近，有可能先接收到k2过期事件，
		// 说明ETCD的计时不是特别精确的。
		// 或者说，ETCD不保证精确的事件顺序。
		// err = etcdStorage.Save("k2", "v2", time.Second*2) 此代码下，有可能先收到k2过期事件

		err = etcdStorage.Save("k2", "v2", time.Second*4)
		err = etcdStorage.Save("k3", "v3", time.Second*8)
		assert.NoError(t, err)

		wc := etcdStorage.Watch()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
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
