package routers

import (
	"user-server-go/internal/handler"
	"user-server-go/internal/token"

	"github.com/gin-gonic/gin"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		loginRouter(group, handler.NewLoginHandler())
	})
}

func loginRouter(group *gin.RouterGroup, h handler.LoginHandler) {

	group.POST("/login", h.Login)                                 // [post] /api/v1/login
	group.POST("/logout", token.GetVerifyHandlerFunc(), h.Logout) // [post] /api/v1/logout

}
