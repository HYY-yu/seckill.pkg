package elastic_job

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/HYY-yu/seckill.pkg/pkg/elastic_job/storage"
)

func TestETCDJob(t *testing.T) {
	cron, err := New(WithStorage(storage.ETCD,
		&storage.Config{
			Endpoints:   []string{"0.0.0.0:2379", "0.0.0.0:12379", "0.0.0.0:22379"},
			DialTimeout: time.Second,
		}))
	assert.NoError(t, err)
	defer func() {
		err = cron.Close()
		assert.NoError(t, err)
	}()
	j := &Job{
		Key:       "test_after",
		DelayTime: time.Now().Add(time.Second * 5).Unix(),
		Cycle:     false,
		Tag:       "TEST",
	}

	t.Run("Normal test", func(t *testing.T) {
		done := make(chan struct{})
		now := time.Now()

		cron.RegisterHandler("TEST", func(j *Job) (err error) {
			t.Log("hello, world! ")
			// 5秒后
			delayTime := time.Unix(j.DelayTime, 0)

			deta := delayTime.Sub(now)
			if deta < 5 {
				t.Errorf("cron time error,now time is %s delayTime is %s ", now.String(), delayTime.String())
			}
			done <- struct{}{}
			return nil
		})

		err = cron.AddJob(j)
		assert.NoError(t, err)

		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		select {
		case <-done:
			return
		case <-ctx.Done():
			t.Errorf("test timeout. ")
		}
	})

	t.Run("test lock", func(t *testing.T) {
		done := make(chan struct{})
		now := time.Now()

		testHander := func(j *Job) (err error) {
			t.Log("hello, world! ")
			// 5秒后
			delayTime := time.Unix(j.DelayTime, 0)

			deta := delayTime.Sub(now)
			if deta < 5 {
				t.Errorf("cron time error,now time is %s delayTime is %s ", now.String(), delayTime.String())
			}
			done <- struct{}{}
			return nil
		}
		cron.RegisterHandler("TEST", testHander)
		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		err = cron.AddJob(j)
		assert.NoError(t, err)

		go func() {
			cron2, err2 := New(WithStorage(storage.ETCD,
				&storage.Config{
					Endpoints:   []string{"0.0.0.0:2379", "0.0.0.0:12379", "0.0.0.0:22379"},
					DialTimeout: time.Second,
				}))
			assert.NoError(t, err2)
			defer func() {
				err = cron2.Close()
				assert.NoError(t, err)
			}()
			cron2.RegisterHandler("TEST", testHander)

			<-ctx.Done()
		}()

		select {
		case <-done:
			<-ctx.Done()
			return
		case <-ctx.Done():
			t.Errorf("test timeout. ")
		}
	})

	t.Run("Multi test", func(t *testing.T) {
		j2 := &Job{
			Key:       "test_2",
			DelayTime: time.Now().Add(time.Second * 10).Unix(),
			Cycle:     false,
			Tag:       "TEST2",
		}
		now := time.Now()

		wg := sync.WaitGroup{}
		wg.Add(2)
		cron.RegisterHandler("TEST", func(j *Job) (err error) {
			// 5秒后
			delayTime := time.Unix(j.DelayTime, 0)
			deta := delayTime.Sub(now)
			if deta < 5 {
				t.Errorf("cron time error,now time is %s delayTime is %s ", now.String(), delayTime.String())
			}
			wg.Done()
			return nil
		})
		cron.RegisterHandler("TEST2", func(j *Job) (err error) {
			// 10秒后
			delayTime := time.Unix(j.DelayTime, 0)
			deta := delayTime.Sub(now)
			if deta < 10 {
				t.Errorf("cron time error,now time is %s delayTime is %s ", now.String(), delayTime.String())
			}
			wg.Done()
			return nil
		})

		err = cron.AddJob(j)
		err = cron.AddJob(j2)
		assert.NoError(t, err)

		wg.Wait()
	})
}

func TestRedisJob(t *testing.T) {
	cron, err := New(WithStorage(storage.Redis,
		&storage.Config{
			Endpoints:   []string{"0.0.0.0:6379"},
			DialTimeout: time.Second,
			Password:    "redis",
		}))

	assert.NoError(t, err)
	defer func() {
		err = cron.Close()
		assert.NoError(t, err)
	}()
	j := &Job{
		Key:       "test_after",
		DelayTime: time.Now().Add(time.Second * 5).Unix(),
		Cycle:     false,
		Tag:       "TEST",
	}

	t.Run("Normal test", func(t *testing.T) {
		done := make(chan struct{})
		now := time.Now()

		cron.RegisterHandler("TEST", func(j *Job) (err error) {
			t.Log("hello, world! ")
			// 5秒后
			delayTime := time.Unix(j.DelayTime, 0)

			deta := delayTime.Sub(now)
			if deta < 5 {
				t.Errorf("cron time error,now time is %s delayTime is %s ", now.String(), delayTime.String())
			}
			done <- struct{}{}
			return nil
		})

		err = cron.AddJob(j)
		assert.NoError(t, err)

		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		select {
		case <-done:
			return
		case <-ctx.Done():
			t.Errorf("test timeout. ")
		}
	})

	t.Run("test lock", func(t *testing.T) {
		done := make(chan struct{})
		now := time.Now()

		testHander := func(j *Job) (err error) {
			t.Log("hello, world! ")
			// time.Sleep(time.Second)
			// 5秒后
			delayTime := time.Unix(j.DelayTime, 0)

			deta := delayTime.Sub(now)
			if deta < 5 {
				t.Errorf("cron time error,now time is %s delayTime is %s ", now.String(), delayTime.String())
			}
			done <- struct{}{}
			return nil
		}
		cron.RegisterHandler("TEST", testHander)
		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		err = cron.AddJob(j)
		assert.NoError(t, err)

		go func() {
			cron2, err2 := New(WithStorage(storage.Redis,
				&storage.Config{
					Endpoints:   []string{"0.0.0.0:6379"},
					DialTimeout: time.Second,
					Password:    "redis",
				}))
			assert.NoError(t, err2)
			defer func() {
				err = cron2.Close()
				assert.NoError(t, err)
			}()
			cron2.RegisterHandler("TEST", testHander)

			<-ctx.Done()
		}()

		select {
		case <-done:
			<-ctx.Done()
			return
		case <-ctx.Done():
			t.Errorf("test timeout. ")
		}
	})

	t.Run("Multi test", func(t *testing.T) {
		j2 := &Job{
			Key:       "test_2",
			DelayTime: time.Now().Add(time.Second * 10).Unix(),
			Cycle:     false,
			Tag:       "TEST2",
		}
		now := time.Now()

		wg := sync.WaitGroup{}
		wg.Add(2)
		cron.RegisterHandler("TEST", func(j *Job) (err error) {
			// 5秒后
			delayTime := time.Unix(j.DelayTime, 0)
			deta := delayTime.Sub(now)
			if deta < 5 {
				t.Errorf("cron time error,now time is %s delayTime is %s ", now.String(), delayTime.String())
			}
			wg.Done()
			return nil
		})
		cron.RegisterHandler("TEST2", func(j *Job) (err error) {
			// 10秒后
			delayTime := time.Unix(j.DelayTime, 0)
			deta := delayTime.Sub(now)
			if deta < 10 {
				t.Errorf("cron time error,now time is %s delayTime is %s ", now.String(), delayTime.String())
			}
			wg.Done()
			return nil
		})

		err = cron.AddJob(j)
		err = cron.AddJob(j2)
		assert.NoError(t, err)

		wg.Wait()
	})
}
