package token

import (
	"user-server-go/internal/cache"
	"user-server-go/internal/database"
	"user-server-go/internal/ecode"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/logger"
)

var tokenCache cache.UserTokenCache

func initTokenCacheConnect() {
	tokenCache = cache.NewUserTokenCache(database.GetCacheType())
}

// Token校验逻辑
func verifyToken(c *gin.Context, tokenString string, claims *Claims) error {
	if tokenCache == nil {
		panic("tokenCache is nil, please call token.Init() first")
	}

	// 查询缓存中 user id
	cacheToken, err := tokenCache.Get(c, claims.UserID)
	if err != nil {
		logger.Warn("get token error", logger.Err(err))
		return ecode.InternalServerError.Err()
	}

	// 缓存中无 user id 情况
	if cacheToken == "" {
		return ecode.ErrNotLogin.Err()
	}

	if cacheToken != tokenString {
		return ecode.Unauthorized.Err()
	}

	return nil
}
