package middleware

import (
	"github.com/HYY-yu/seckill.pkg/core"
	"github.com/juju/ratelimit"
	"go.uber.org/zap"

	"github.com/HYY-yu/seckill.pkg/pkg/response"
)

var _ Middleware = (*middleware)(nil)

type Middleware interface {
	// Jwt 中间件
	Jwt(ctx core.Context) (userId int64, userName string, err response.Error)

	// DisableLog 不记录日志
	DisableLog() core.HandlerFunc

	RequestLimit() core.HandlerFunc
}

type middleware struct {
	logger *zap.Logger

	jwtSecret string

	bucket *ratelimit.Bucket
}

func New(logger *zap.Logger, jwtSecret string) Middleware {
	return &middleware{
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

func NewWithLimiter(logger *zap.Logger, jwtSecret string, rate float64, cap int64) Middleware {
	bucket := ratelimit.NewBucketWithRate(rate, cap)
	return &middleware{
		logger:    logger,
		jwtSecret: jwtSecret,
		bucket:    bucket,
	}
}
