package elastic_job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/HYY-yu/seckill.pkg/pkg/elastic_job/storage"
	"github.com/HYY-yu/seckill.pkg/pkg/encrypt"
)

// ElasticJob 分布式定时（延时）执行器
type ElasticJob interface {
	// AddJob 添加新任务，此任务将发送到存储器中进行时间计算，并监听存储器的回调事件
	AddJob(ctx context.Context, j *Job) error
	// RegisterHandler 收到回调事件，执行回调函数链
	// 利用ETCD做分布式锁，保证回调链只有一个能被执行。
	RegisterHandler(handlerTag string, h Handler)

	Close() error
}

type config struct {
	storageType   storage.Type
	storageConfig *storage.Config

	logger *zap.Logger

	shouldMetrics bool
	serverName    string
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

func WithMetrics() Options {
	return func(c *config) {
		c.shouldMetrics = true
	}
}

func WithServerName(serverName string) Options {
	return func(c *config) {
		c.serverName = serverName
	}
}

type elasticJob struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config

	store storage.BackendStorage

	handlers *sync.Map

	logger *zap.Logger

	metrics *JobMetrics
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

	if cfg.shouldMetrics {
		cron.metrics = NewJobMetrics("")
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
			respJob, err := UnmarshalJson(wresp.Value)
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
				ts := time.Now()

				err = hander.(Handler)(respJob)
				if err != nil {
					e.logger.Error("handler report error",
						zap.String("key", wresp.Key),
						zap.String("tag", respJob.Tag),
						zap.Error(err),
					)
				}

				costSeconds := time.Since(ts).Seconds()
				if e.cfg.shouldMetrics {
					e.metrics.MetricsRunCost(e.cfg.serverName, respJob.Key, costSeconds)
				}

				time.Sleep(time.Second * 3)
				// 锁不能马上释放，如果handler执行的太快，锁就马上被释放了。
				// 导致别的节点也能获取到锁，只有充分的锁住足够的时间，让过期事件被全量推送
				// 其它节点已经退出，才可以释放锁。
				err = e.store.UnLock(jobHash)
				if err != nil {
					e.logger.Error("handler unlock error ",
						zap.Error(err),
					)
				}
			}()

			if respJob.Cycle {
				// 再次写入，注意，循环任务无可用性保证。
				// 有可能会中断循环
				err = e.AddJob(e.ctx, respJob)
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

func (e *elasticJob) AddJob(ctx context.Context, j *Job) error {
	value := j.MarshalJson()
	delay := time.Until(time.Unix(j.DelayTime, 0))
	if delay <= 0 {
		return fmt.Errorf("the delay_time must happen in the future. ")
	}

	err := e.store.Save(j.Key, value, delay)
	if err != nil {
		return err
	}
	if e.cfg.shouldMetrics {
		span := trace.SpanFromContext(ctx)
		var traceId string
		if span.SpanContext().HasTraceID() {
			traceId = span.SpanContext().TraceID().String()
		}
		e.metrics.MetricsAddTotal(e.cfg.serverName, j.Key, traceId)
	}

	return nil
}

func (e *elasticJob) RegisterHandler(handlerTag string, h Handler) {
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
