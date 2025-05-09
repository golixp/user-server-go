package token

import (
	"time"
	"user-server-go/internal/cache"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/errcode"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/go-dev-frame/sponge/pkg/logger"
)

const (
	headerAuthorizationKey = "Authorization"

	contextTokenKey  = "token"
	conetxtClaimsKey = "claims"
)

func GetVerifyHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader(headerAuthorizationKey)
		if len(authorization) < 100 {
			response.Out(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		tokenString := authorization[7:] // remove Bearer prefix

		claims, err := ParseJwtToken(tokenString)
		if err != nil {
			response.Out(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		c.Set(contextTokenKey, tokenString)
		c.Set(conetxtClaimsKey, claims)

		if err = verifyToken(c, tokenString, claims); err != nil {
			response.Out(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		if err = autoRenewToken(c, claims); err != nil {
			logger.Warn("renew token fail ", logger.Err(err))
		}

		c.Next()
	}
}

func autoRenewToken(c *gin.Context, claims *Claims) error {

	// 临近10分钟过期时重新续签 token
	if now := time.Now(); claims.ExpiresAt.Unix()-now.Unix() < int64(time.Minute*10) {

		newClaims := NewClaims(claims.UserID)
		newToken, err := newClaims.GenerateJwtToken()
		if err != nil {
			return err
		}

		// 重设 token 缓存
		tokenCache.Set(c, claims.UserID, newToken, cache.UserTokenExpireTime)

		// 重设 gin context 信息
		c.Set(contextTokenKey, newToken)
		c.Set(conetxtClaimsKey, newClaims)

		// 配置 header 信息
		c.Header("X-Renewed-Token", newToken)
	}

	return nil
}

func GetTokenFromCtx(c *gin.Context) (token string, ok bool) {
	tokenValue, exists := c.Get(contextTokenKey)
	if !exists {
		return "", false
	}

	token, ok = tokenValue.(string)

	return
}

func GetClaimsFromCtx(c *gin.Context) (claims *Claims, ok bool) {
	claimsVlaue, exists := c.Get(conetxtClaimsKey)
	if !exists {
		return nil, false
	}

	claims, ok = claimsVlaue.(*Claims)

	return
}
