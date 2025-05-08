package token

import (
	"user-server-go/internal/config"
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
