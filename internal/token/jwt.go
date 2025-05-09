package token

import (
	"fmt"
	"time"
	"user-server-go/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSignKey []byte

func initJwtSignKey() {
	jwtSignKey = []byte(config.Get().App.JwtSignKey)
}

func GetJwtSignKey() []byte {
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

// NewClaimsWithUidAndExp 根据用户ID和过期时间生成Claims
func NewClaimsWithUidAndExp(uid uint64, expDuration time.Duration) *Claims {
	claims := &Claims{
		UserID: uid,
	}
	claims.SetExpAndIat(expDuration)

	return claims
}

// GenerateJWTToken 生成一个新的JWT令牌。
func GenerateJwtToken(claims Claims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(GetJwtSignKey())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseJWTToken 解析并验证JWT令牌。
func ParseJWTToken(tokenString string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return GetJwtSignKey(), nil
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
