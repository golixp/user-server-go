package token

import (
	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/errcode"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
)

func GetVerifyHandlerFunc() gin.HandlerFunc {
	verifyFunc := VerifyToken
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
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

		c.Set("token", tokenString)
		c.Set("claims", claims)

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
