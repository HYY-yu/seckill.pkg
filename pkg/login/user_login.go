package login

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HYY-yu/werror"
	"github.com/go-redis/redis/v8"

	"github.com/HYY-yu/seckill.pkg/cache_v2"
	"github.com/HYY-yu/seckill.pkg/pkg/login/model"
	"github.com/HYY-yu/seckill.pkg/pkg/token"
)

// 通用用户登录、登出凭证管理系统

// LoginTokenSystem
// 负责Token生成、失效、刷新
// 不负责Token的校验，Token校验通过JWT的签名机制
// 凭证系统职责：
// 1. 提供凭证，并且支持校验凭证
// 2. 凭证可主动失效（客户端、服务端均可主动）
// 3. 凭证过期可刷新
type LoginTokenSystem interface {
	// GenerateToken 生成Token，保证Token的可校验性（JWT）
	GenerateToken(ctx context.Context, userId int, userName string) (*model.LoginResponse, error)
	// TokenCancelById 可根据 userId \ userName 失效放出的Token
	TokenCancelById(ctx context.Context, userId int, userName string) error
	// TokenCancel 可根据 token 本身失效放出的 token
	TokenCancel(ctx context.Context, token string) error
	// RefreshToken 失效 oldToken (或者oldToken已经过期，自动失效)，发放新 Token
	RefreshToken(ctx context.Context, oldToken string) (*model.LoginResponse, error)
}

func NewByRefreshToken(cfg *RefreshTokenConfig, cache cache_v2.Repo) LoginTokenSystem {
	return &RefreshTokenSystem{cfg: cfg, cache: cache}
}

type RefreshTokenConfig struct {
	Secret          string        `json:"secret"`
	ExpireDuration  time.Duration `json:"expire_duration"`
	RefreshDuration time.Duration `json:"refresh_duration"`
}

type RefreshTokenSystem struct {
	cfg *RefreshTokenConfig

	cache cache_v2.Repo
}

func (r *RefreshTokenSystem) GenerateToken(ctx context.Context, userId int, userName string) (*model.LoginResponse, error) {
	accessToken, err := token.New(r.cfg.Secret).JwtSign(int64(userId), userName, r.cfg.ExpireDuration)
	if err != nil {
		return nil, err
	}

	refreshToken := r.generateRefreshToken(r.cfg.Secret, userId, userName)

	userClaims := struct {
		UserId   int    `json:"user_id"`
		UserName string `json:"user_name"`
	}{
		UserId:   userId,
		UserName: userName,
	}

	userJson, _ := json.Marshal(userClaims)

	err = r.cache.Set(ctx, model.RedisRefreshTokenKeyPrefix+refreshToken, string(userJson), r.cfg.RefreshDuration)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token: &model.LoginResponseByRefreshToken{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

func (r *RefreshTokenSystem) TokenCancelById(ctx context.Context, userId int, userName string) error {
	refreshToken := r.generateRefreshToken(r.cfg.Secret, userId, userName)
	return r.TokenCancel(ctx, refreshToken)
}

func (r *RefreshTokenSystem) TokenCancel(ctx context.Context, refreshToken string) error {
	_ = r.cache.Del(ctx, model.RedisRefreshTokenKeyPrefix+refreshToken)
	return nil
}

func (r *RefreshTokenSystem) RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, error) {
	userJson, err := r.cache.Get(ctx, model.RedisRefreshTokenKeyPrefix+refreshToken)
	if err != nil {
		if werror.Is(err, redis.Nil) {
			return nil, model.RefreshTokenExpired
		}
		return nil, err
	}

	userClaims := struct {
		UserId   int    `json:"user_id"`
		UserName string `json:"user_name"`
	}{}
	_ = json.Unmarshal([]byte(userJson), &userClaims)

	// 删除原refreshToken
	_ = r.TokenCancel(ctx, refreshToken)

	return r.GenerateToken(ctx, userClaims.UserId, userClaims.UserName)
}

func (u *RefreshTokenSystem) generateRefreshToken(secret string, userId int, userName string) string {
	// RefreshToken = hmac256(accessToken,jwtConfig.Secret)
	hencrypt := hmac.New(md5.New, []byte(secret))
	hencrypt.Write([]byte(fmt.Sprintf("%d_%s", userId, userName)))
	return fmt.Sprintf("%x", hencrypt.Sum(nil))
}

func NewByBlackList(cfg *BlackListConfig, cache cache_v2.Repo) LoginTokenSystem {
	return &BlackListSystem{cfg: cfg, cache: cache}
}

type BlackListConfig struct {
	Secret         string        `json:"secret"`
	ExpireDuration time.Duration `json:"expire_duration"`
}

type BlackListSystem struct {
	cfg *BlackListConfig

	cache cache_v2.Repo
}

func (r *BlackListSystem) GenerateToken(ctx context.Context, userId int, userName string) (*model.LoginResponse, error) {
	accessToken, err := token.New(r.cfg.Secret).JwtSign(int64(userId), userName, r.cfg.ExpireDuration)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token: &model.LoginResponseByBlackList{
			AccessToken: accessToken,
		},
	}, nil
}

func (r *BlackListSystem) blackListKey(userId int64, userName string) string {
	return fmt.Sprintf("%s_%d_%s", model.RedisBlackListKeyPrefix, userId, userName)
}

// CheckBlackList 验证时，需要验证是否在黑名单
func (r *BlackListSystem) CheckBlackList(ctx context.Context, accessToken string) (bool, error) {
	claim, err := token.New(r.cfg.Secret).JwtParseUnsafe(accessToken)
	if err != nil {
		return false, fmt.Errorf("this token is unvalid ")
	}
	key := r.blackListKey(claim.UserID, claim.UserName)

	a, err := r.cache.Get(ctx, key)
	if err != nil {
		if werror.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return a == "1", nil
}

// TokenCancelById 根据 userId 失效 Token
func (r *BlackListSystem) TokenCancelById(ctx context.Context, userId int, userName string) error {
	key := r.blackListKey(int64(userId), userName)

	err := r.cache.Set(ctx, key, "1", r.cfg.ExpireDuration)
	if err != nil {
		return err
	}
	return nil
}

func (r *BlackListSystem) TokenCancel(ctx context.Context, accessToken string) error {
	claim, err := token.New(r.cfg.Secret).JwtParse(accessToken)
	if err != nil {
		// 解析失败，无需放到黑名单，这个token校验不过。
		return nil
	}

	// 获取 claim 中的 userId+userName
	// 进入黑名单的是这个人，所以他相关的所有AccessToken都会失效
	key := r.blackListKey(claim.UserID, claim.UserName)
	err = r.cache.Set(ctx, key, "1", r.cfg.ExpireDuration)
	if err != nil {
		return err
	}
	return nil
}

// RefreshToken 刷新无法校验oldToken的正确性
func (r *BlackListSystem) RefreshToken(ctx context.Context, accessToken string) (*model.LoginResponse, error) {
	claim, err := token.New(r.cfg.Secret).JwtParseUnsafe(accessToken)
	if err != nil {
		return nil, fmt.Errorf("this token is unvalid ")
	}
	return r.GenerateToken(ctx, int(claim.UserID), claim.UserName)
}
