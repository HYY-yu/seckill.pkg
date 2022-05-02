package elastic_job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/HYY-yu/seckill.pkg/pkg/elastic_job/job"
	"github.com/HYY-yu/seckill.pkg/pkg/elastic_job/storage"
	"github.com/HYY-yu/seckill.pkg/pkg/encrypt"
)

// ElasticJob 分布式定时（延时）执行器
type ElasticJob interface {
	// AddJob 添加新任务，此任务将发送到存储器中进行时间计算，并监听存储器的回调事件
	AddJob(j *job.Job) error
	// RegisterHandler 收到回调事件，执行回调函数链
	// 利用ETCD做分布式锁，保证回调链只有一个能被执行。
	RegisterHandler(handlerTag string, h job.Handler)

	Close() error
}

type config struct {
	storageType   storage.Type
	storageConfig *storage.Config

	logger *zap.Logger
}

type Options func(c *config)

func WithStorage(t storage.Type, sc *storage.Config) Options {
	return func(c *config) {
		c.storageType = t
		c.storageConfig = sc
	}
}

func WithLogger(l *zap.Logger) Options {
	return func(c *config) {
		c.logger = l
	}
}

type elasticJob struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config

	store storage.BackendStorage

	handlers *sync.Map

	logger *zap.Logger
}

func New(opts ...Options) (ElasticJob, error) {
	var err error
	cfg := &config{}

	for _, opt := range opts {
		opt(cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cron := &elasticJob{
		ctx:      ctx,
		cancel:   cancel,
		cfg:      cfg,
		handlers: &sync.Map{},
	}
	if cfg.logger != nil {
		cron.logger = cfg.logger
	} else {
		cron.logger, err = zap.NewDevelopment()
		if cron.logger != nil {
			cron.logger.Named("elastic-cron")
		}
	}
	if err != nil {
		return nil, err
	}

	// init storage
	switch cfg.storageType {
	case storage.ETCD:
		cron.store, err = storage.NewEtcdStorage(cfg.storageConfig)
	case storage.Redis:
		cron.store, err = storage.NewRedisStorage(cfg.storageConfig)
	default:
		cron.store, err = storage.NewEtcdStorage(cfg.storageConfig)
	}
	if err != nil {
		return nil, fmt.Errorf("storage init error: %w ", err)
	}
	go cron.run()
	return cron, nil
}

func (e *elasticJob) run() {
	defer func() {
		if err := recover(); err != nil {
			go e.run()
		}
	}()

	wc := e.store.Watch()
	for {
		select {
		case wresp := <-wc:
			e.logger.Sugar().Infof("a notification event: %s ", wresp.Key)
			respJob, err := job.UnmarshalJson(wresp.Value)
			if err != nil {
				e.logger.Error("cannot unmarshal job value",
					zap.String("key", wresp.Key),
					zap.String("value", wresp.Value),
					zap.Int64("timestamp", wresp.TimeNow),
					zap.Error(err),
				)
			}
			hander, ok := e.handlers.Load(respJob.Tag)
			if !ok {
				// 没有注册
				e.logger.Sugar().Warnf("not found hander for job tag: %s ", respJob.Tag)
				continue
			}

			// 异步执行任务
			go func() {
				jobHash := encrypt.MD5(respJob.Key + respJob.Tag)
				err = e.store.TryLock(jobHash)
				if err != nil {
					if err == storage.ErrLocked {
						// 被别的节点申请到，直接退出
						e.logger.Sugar().Info("the job has already running. ")
						return
					}
					e.logger.Error("handler lock error ",
						zap.Error(err),
					)
					return
				}

				fmt.Println("got lock")
				err = hander.(job.Handler)(respJob)
				if err != nil {
					e.logger.Error("handler report error",
						zap.String("key", wresp.Key),
						zap.String("tag", respJob.Tag),
						zap.Error(err),
					)
				}

				err = e.store.UnLock(jobHash)
				if err != nil {
					e.logger.Error("handler unlock error ",
						zap.Error(err),
					)
				}
				fmt.Println("release lock")
			}()

			if respJob.Cycle {
				// 再次写入，注意，循环任务无可用性保证。
				// 有可能会中断循环
				err = e.AddJob(respJob)
				if err != nil {
					e.logger.Error("deal cycle job error ",
						zap.String("key", wresp.Key),
						zap.String("tag", respJob.Tag),
						zap.Error(err),
					)
				}
			}
		case <-e.ctx.Done():
			// 关闭
			return
		}
	}
}

func (e *elasticJob) AddJob(j *job.Job) error {
	value := j.MarshalJson()
	delay := time.Until(time.Unix(j.DelayTime, 0))
	if delay <= 0 {
		return fmt.Errorf("the delay_time must happen in the future. ")
	}

	return e.store.Save(j.Key, value, delay)
}

func (e *elasticJob) RegisterHandler(handlerTag string, h job.Handler) {
	e.handlers.Store(handlerTag, h)
}

func (e *elasticJob) Close() error {
	e.cancel()
	_ = e.logger.Sync()

	err := e.store.Close()
	if err != nil {
		return err
	}
	return nil
}
