package authorize

import (
	"crypto/rsa"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	Role     string `json:"role"`
}

type FacilityClaims struct {
	jwt.RegisteredClaims
	UserFacilities []string `json:"userFacilities"`
}
type AuthHandlerInterface interface {
	GetPublicKey() *rsa.PublicKey
	GetFacilityTokenPublicKey() *rsa.PublicKey
	GetAccessToken(c *gin.Context) (*CustomClaims, error)
	GetFacilityToken(c *gin.Context) (*FacilityClaims, error)
}
