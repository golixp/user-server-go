package token

import (
	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware"
)

func GetVerifyHandlerFunc() gin.HandlerFunc {
	return middleware.Auth(
		middleware.WithSignKey(GetJwtSignKey()),
		middleware.WithExtraVerify(VerifyToken),
	)
}
