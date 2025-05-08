package token

import (
	"strconv"
	"user-server-go/internal/cache"
	"user-server-go/internal/database"
	"user-server-go/internal/ecode"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/jwt"
	"github.com/go-dev-frame/sponge/pkg/logger"
)

var tokenCache cache.UserTokenCache

func initTokenCacheConnect() {
	tokenCache = cache.NewUserTokenCache(database.GetCacheType())
}

// Token校验逻辑
func VerifyToken(claims *jwt.Claims, c *gin.Context) error {
	if tokenCache == nil {
		panic("tokenCache is nil, please call token.Init() first")
	}

	// 获取 user id
	uid, err := strconv.ParseUint(claims.UID, 10, 64)
	if err != nil {
		return err
	}

	// 查询缓存中 user id
	token, err := tokenCache.Get(c, uid)
	if err != nil {
		logger.Warn("get token error", logger.Err(err))
		return ecode.InternalServerError.Err()
	}

	// 缓存中无 user id 情况
	if token == "" {
		return ecode.ErrNotLogin.Err()
	}

	// 生成缓存中 Token 的 Claims
	cacheClaims, err := jwt.ValidateToken(token, jwt.WithValidateTokenSignKey(GetJwtSignKey()))
	if err != nil {
		return err
	}

	// 校验 user id 一致性
	if cacheClaims.UID != claims.UID {
		return ecode.Unauthorized.Err()
	}

	return nil
}
