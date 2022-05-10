package model

import (
	"errors"
)

type LoginParams struct {
	UserName string `form:"user_name" v:"required|length:3,20"`
}

type LoginResponse struct {
	Token interface{} `json:"token"` // LoginResponseByRefreshToken or LoginResponseByBlackList
}

const (
	RedisRefreshTokenKeyPrefix = "sk:refresh:"
	RedisBlackListKeyPrefix    = "sk:black_list:"
)

type LoginResponseByRefreshToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginResponseByBlackList struct {
	AccessToken string `json:"access_token"`
}

var (
	RefreshTokenExpired = errors.New("the refresh token is expired. ")
)
