package authorize

import (
	"crypto/rsa"
	"stock/db"
	"stock/myRedis"
)

type AuthHandlerInterface interface {
	GetPublicKey() *rsa.PublicKey
}

type authHandler struct {
	publicKey *rsa.PublicKey
	// インターフェースから定義してモックが作れるような設計する
	db    db.UserDB
	redis myRedis.RedisPool
}

func (a *authHandler) GetPublicKey() *rsa.PublicKey {
	return a.publicKey
}

func NewAuthHandler(publicKey *rsa.PublicKey, db db.UserDB, redis myRedis.RedisPool) AuthHandlerInterface {
	return &authHandler{
		publicKey: publicKey,
		db:        db,
		redis:     redis,
	}
}
