package token

import (
	"fmt"
	"time"
	"user-server-go/internal/cache"
	"user-server-go/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSignKey []byte

func initJwtSignKey() {
	jwtSignKey = []byte(config.Get().App.JwtSignKey)
}

func getJwtSignKey() []byte {
	if jwtSignKey == nil {
		panic("jwtKey not initialized")
	}
	return jwtSignKey
}

type Claims struct {
	UserID uint64         `json:"uid,omitempty"`    // 用户ID
	Fields map[string]any `json:"fields,omitempty"` // custom fields
	jwt.RegisteredClaims
}

// SetExpAndIat 根据过期间隔自动设置过期时间和签发时间, 签发时间为当前时间
func (c *Claims) SetExpAndIat(expDuration time.Duration) {
	now := time.Now()
	exp := now.Add(expDuration)
	c.IssuedAt = jwt.NewNumericDate(now)
	c.ExpiresAt = jwt.NewNumericDate(exp)
}

// GenerateJWTToken 生成一个新的JWT令牌。
func (c *Claims) GenerateJwtToken() (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	tokenString, err := token.SignedString(getJwtSignKey())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// NewClaims 根据用户ID和过期时间生成Claims
func NewClaims(uid uint64) *Claims {
	claims := &Claims{
		UserID: uid,
	}
	expDuration := cache.UserTokenExpireTime
	claims.SetExpAndIat(expDuration)

	return claims
}

// ParseJwtToken 解析并验证JWT令牌。
func ParseJwtToken(tokenString string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJwtSignKey(), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid JWT token or claims")
	}

	return claims, nil
}
