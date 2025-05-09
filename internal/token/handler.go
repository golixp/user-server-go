package token

import (
	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/errcode"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
)

const (
	headerAuthorizationKey = "Authorization"

	contextTokenKey  = "token"
	conetxtClaimsKey = "claims"
)

func GetVerifyHandlerFunc() gin.HandlerFunc {
	verifyFunc := VerifyToken
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

		if verifyFunc != nil {
			if err = verifyFunc(c, tokenString, claims); err != nil {
				response.Out(c, errcode.Unauthorized)
				c.Abort()
				return
			}
		}

		c.Next()
	}
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
