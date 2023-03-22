package middleware

import (
	"errors"
	"github.com/HYY-yu/seckill.pkg/core"
	"github.com/HYY-yu/seckill.pkg/pkg/response"
	"net/http"
)

func (m *middleware) RequestLimit() core.HandlerFunc {
	return func(c core.Context) {
		if m.bucket != nil && m.bucket.TakeAvailable(1) < 1 {
			err := response.NewErrorAutoMsg(
				http.StatusForbidden,
				response.TooManyRequests,
			).WithErr(errors.New("限流"))
			c.AbortWithError(err)
			return
		}
	}
}
