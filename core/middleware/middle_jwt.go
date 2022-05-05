package middleware

import (
	"errors"
	"net/http"

	"github.com/HYY-yu/seckill.pkg/core"
	"github.com/HYY-yu/seckill.pkg/pkg/response"
	"github.com/HYY-yu/seckill.pkg/pkg/token"
)

func (m *middleware) Jwt(ctx core.Context) (userId int64, userName string, err response.Error) {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		err = response.NewErrorAutoMsg(
			http.StatusUnauthorized,
			response.AuthorizationError,
		).WithErr(errors.New("Header 中缺少 Authorization 参数 "))
		return
	}

	claims, errParse := token.New(m.jwtSecret).JwtParse(auth)
	if errParse != nil {
		err = response.NewErrorAutoMsg(
			http.StatusUnauthorized,
			response.AuthorizationError,
		).WithErr(errParse)

		return
	}

	userId = claims.UserID
	if userId <= 0 {
		err = response.NewErrorAutoMsg(
			http.StatusUnauthorized,
			response.AuthorizationError,
		).WithErr(errors.New("claims.UserID <= 0 "))

		return
	}
	userName = claims.UserName
	return
}
