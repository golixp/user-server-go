package handler

import (
	"context"
	"errors"
	"strconv"
	"time"
	"user-server-go/internal/cache"
	"user-server-go/internal/dao"
	"user-server-go/internal/database"
	"user-server-go/internal/ecode"
	"user-server-go/internal/model"
	"user-server-go/internal/token"
	"user-server-go/internal/types"
	"user-server-go/pkg/password"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/go-dev-frame/sponge/pkg/jwt"
	"github.com/go-dev-frame/sponge/pkg/logger"
)

var _ LoginHandler = (*loginHandler)(nil)

// LoginHandler defining the handler interface
type LoginHandler interface {
	Login(c *gin.Context)
	Logout(c *gin.Context)
}

type loginHandler struct {
	iDao       dao.UserDao
	tokenCache cache.UserTokenCache
}

// NewLoginHandler creating the handler interface
func NewLoginHandler() LoginHandler {
	return &loginHandler{
		iDao: dao.NewUserDao(
			database.GetDB(), // db driver is sqlite
			cache.NewUserCache(database.GetCacheType()),
		),
		tokenCache: cache.NewUserTokenCache(database.GetCacheType()),
	}
}

// 登录
// @Summary login
// @Description submit information to login
// @Tags login
// @accept json
// @Produce json
// @Param data body types.LoginRequest true "login information"
// @Success 200 {object} types.LoginReply{}
// @Router /api/v1/login [post]
// @Security BearerAuth
func (h *loginHandler) Login(c *gin.Context) {
	form := &types.LoginRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	// 生成 context.Context
	ctx := middleware.WrapCtx(c)

	// 搜索用户
	user, err := h.iDao.GetByUsername(ctx, form.Username)
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			response.Error(c, ecode.ErrUserNotExists)
		} else {
			logger.Warn("login GetByUsername error", logger.Err(err), logger.String("username", form.Username), middleware.CtxRequestIDField(c))
			response.Output(c, ecode.InternalServerError.ToHTTPCode())
			return
		}
	}

	// 验证密码
	if !password.VerifyPassword(form.Password, user.Password) {
		response.Error(c, ecode.ErrPassword)
		return
	}

	// 生成 JWT Token
	claims := token.NewClaimsWithUidAndExp(user.ID, cache.UserTokenExpireTime)
	token, err := claims.GenerateJwtToken()
	if err != nil {
		logger.Error("GenerateToken error", logger.Err(err), middleware.CtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	// 缓存Token
	err = h.tokenCache.Set(c, user.ID, token, cache.UserTokenExpireTime)
	if err != nil {
		logger.Error("h.userTokenCache.Set error", logger.Err(err), middleware.CtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	// 更新登录相关状态
	userInfo := &model.User{
		LoginAt: time.Now(),
		LoginIP: c.ClientIP(),
	}
	userInfo.ID = user.ID
	err = h.iDao.UpdateByID(ctx, userInfo)
	if err != nil {
		logger.Warn("h.iDao.UpdateByID error", logger.Err(err), logger.Any("user", userInfo), middleware.CtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	resp := &types.TokenObjDetail{
		ID:    user.ID,
		Token: token,
	}
	response.Success(c, resp)
}

// 退出登录
// @Summary logout
// @Description submit information to logout
// @Tags login
// @accept json
// @Produce json
// @Param data body types.LogoutRequest true "logout information"
// @Success 200 {object} types.LogoutReply{}
// @Router /api/v1/logout [post]
// @Security BearerAuth
func (h *loginHandler) Logout(c *gin.Context) {
	form := &types.LogoutRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	// 获取 gin.Context 中的Token信息
	claims, ok := middleware.GetClaims(c)
	if !ok {
		logger.Error("GetClaims error", middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	// 获取 user id
	uid, err := strconv.ParseUint(claims.UID, 10, 64)
	if err != nil {
		logger.Error("strconv error", logger.String("uid", claims.UID), logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	// 获取缓存中的 Token
	cacheToken, err := h.getLoginToken(c, uid)
	if errors.Is(err, ecode.ErrNotLogin.Err()) {
		response.Error(c, ecode.ErrNotLogin)
		return
	}
	if err != nil {
		logger.Error("checkLogin error", logger.Err(err), middleware.CtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	// 检查ID一致性
	cacheClaims, err := jwt.ValidateToken(cacheToken, jwt.WithValidateTokenSignKey(token.GetJwtSignKey()))
	if err != nil {
		logger.Error("ValidateToken error", logger.String("token", cacheToken), logger.Err(err), middleware.CtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}
	if claims.UID != cacheClaims.UID {
		response.Error(c, ecode.ErrNotLogin)
		return
	}

	// 删除缓存
	err = h.tokenCache.Del(c, uid)
	if err != nil {
		logger.Error("TokenCache.Del error", logger.Uint64("uid", uid), logger.Err(err), middleware.CtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)

}

// CheckLoginToken 获取指定登录用户缓存的Token
func (h *loginHandler) getLoginToken(c context.Context, id uint64) (string, error) {
	token, err := h.tokenCache.Get(c, id)
	if err != nil {
		logger.Warn("get token error", logger.Err(err))
		return "", ecode.InternalServerError.Err()
	}

	if token == "" {
		return "", ecode.ErrNotLogin.Err()
	}

	return token, nil
}
