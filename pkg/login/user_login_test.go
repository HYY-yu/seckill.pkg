package login

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/HYY-yu/seckill.pkg/cache_v2"
	"github.com/HYY-yu/seckill.pkg/pkg/login/model"
	"github.com/HYY-yu/seckill.pkg/pkg/token"
)

func TestRefreshToken(t *testing.T) {
	cacheRepo, err := cache_v2.New("test", &cache_v2.RedisConf{
		Addr: "127.0.0.1:6379",
		Pass: "redis",
	})
	assert.NoError(t, err)

	cfg := &RefreshTokenConfig{
		Secret:          "test_secret",
		ExpireDuration:  time.Second * 2,
		RefreshDuration: time.Second * 10,
	}
	ctx := context.Background()
	userId := 1
	userName := "UserName"

	system := NewByRefreshToken(cfg, cacheRepo)

	t.Run("Test token expired", func(t *testing.T) {
		resp, err := system.GenerateToken(ctx, userId, userName)
		assert.NoError(t, err)

		refreshResp := resp.Token.(*model.LoginResponseByRefreshToken)
		t.Logf("Generate accesstoken: %s", refreshResp.AccessToken)
		t.Logf("Generate refreshtoken: %s", refreshResp.RefreshToken)

		claims, err := token.New(cfg.Secret).JwtParse(refreshResp.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, claims.UserID, int64(userId))
		assert.Equal(t, claims.UserName, userName)

		time.Sleep(time.Second * 3)

		_, err = token.New(cfg.Secret).JwtParse(refreshResp.AccessToken)
		assert.Equal(t, token.ErrorTokenExpiredOrNotActive, err)
	})

	t.Run("Test token refresh", func(t *testing.T) {
		resp, err := system.GenerateToken(ctx, userId, userName)
		assert.NoError(t, err)

		refreshResp := resp.Token.(*model.LoginResponseByRefreshToken)
		time.Sleep(time.Second * 3)

		newResp, err := system.RefreshToken(ctx, refreshResp.RefreshToken)
		assert.NoError(t, err)
		newRefreshResp := newResp.Token.(*model.LoginResponseByRefreshToken)

		// 依然有效
		claims, err := token.New(cfg.Secret).JwtParse(newRefreshResp.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, claims.UserID, int64(userId))
		assert.Equal(t, claims.UserName, userName)
	})

	t.Run("Test token cancel", func(t *testing.T) {
		resp, err := system.GenerateToken(ctx, userId, userName)
		assert.NoError(t, err)

		refreshResp := resp.Token.(*model.LoginResponseByRefreshToken)

		err = system.TokenCancel(ctx, refreshResp.RefreshToken)
		assert.NoError(t, err)

		// 无法刷新
		_, err = system.RefreshToken(ctx, refreshResp.RefreshToken)
		assert.Equal(t, model.RefreshTokenExpired, err)
	})
}

func TestBlackList(t *testing.T) {
	cacheRepo, err := cache_v2.New("test", &cache_v2.RedisConf{
		Addr: "127.0.0.1:6379",
		Pass: "redis",
	})
	assert.NoError(t, err)

	cfg := &BlackListConfig{
		Secret:         "test_secret",
		ExpireDuration: time.Second * 2,
	}
	ctx := context.Background()
	userId := 1
	userName := "UserName"

	system := NewByBlackList(cfg, cacheRepo)

	t.Run("Test token expired", func(t *testing.T) {
		resp, err := system.GenerateToken(ctx, userId, userName)
		assert.NoError(t, err)

		refreshResp := resp.Token.(*model.LoginResponseByBlackList)
		t.Logf("Generate accesstoken: %s", refreshResp.AccessToken)

		claims, err := token.New(cfg.Secret).JwtParse(refreshResp.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, claims.UserID, int64(userId))
		assert.Equal(t, claims.UserName, userName)

		time.Sleep(time.Second * 3)

		_, err = token.New(cfg.Secret).JwtParse(refreshResp.AccessToken)
		assert.Equal(t, token.ErrorTokenExpiredOrNotActive, err)
	})

	t.Run("Test token refresh", func(t *testing.T) {
		resp, err := system.GenerateToken(ctx, userId, userName)
		assert.NoError(t, err)

		refreshResp := resp.Token.(*model.LoginResponseByBlackList)

		newResp, err := system.RefreshToken(ctx, refreshResp.AccessToken)
		assert.NoError(t, err)

		time.Sleep(time.Second * 4)

		newRefreshResp := newResp.Token.(*model.LoginResponseByBlackList)

		// 依然有效
		claims, err := token.New(cfg.Secret).JwtParse(newRefreshResp.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, claims.UserID, int64(userId))
		assert.Equal(t, claims.UserName, userName)

		// 无法让老的AccessToken过期，只能等待它自然过期
	})

	t.Run("Test token cancel", func(t *testing.T) {
		resp, err := system.GenerateToken(ctx, userId, userName)
		assert.NoError(t, err)

		refreshResp := resp.Token.(*model.LoginResponseByBlackList)

		// 不在黑名单
		result, err := system.(*BlackListSystem).CheckBlackList(ctx, refreshResp.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, false, result)

		// 进入黑名单
		err = system.TokenCancel(ctx, refreshResp.AccessToken)
		assert.NoError(t, err)

		result, err = system.(*BlackListSystem).CheckBlackList(ctx, refreshResp.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, true, result)
	})
}
