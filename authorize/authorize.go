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

type Role struct {
	HotelID string `json:"HotelID"`
	Role    string `json:"Role"`
}

type HotelRoles string

const (
	StatusOwner        HotelRoles = "owner"
	StatusManager      HotelRoles = "manager"
	StatusFront        HotelRoles = "front"
	StatusHousekeeping HotelRoles = "housekeeping"
	StatusStaff        HotelRoles = "staff"
)

// Custom claim to permitted FacilityInfo
type FacilityClaims struct {
	jwt.RegisteredClaims
	UserFacilities []Role `json:"userFacilities"`
}
type AuthHandlerInterface interface {
	GetPublicKey() *rsa.PublicKey
	GetFacilityTokenPublicKey() *rsa.PublicKey
	GetAccessToken(c *gin.Context) (*CustomClaims, error)
	GetFacilityToken(c *gin.Context) (*FacilityClaims, error)
}
