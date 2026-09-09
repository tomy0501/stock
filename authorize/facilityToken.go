package authorize

import (
	"crypto/rsa"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type FacilityClaim struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

type FacilityAuthData struct {
	FacilityID string
	UserID     string
}

type Role string

const (
	OwnerRole        Role = "owner"
	ManagerRole      Role = "manager"
	StaffRole        Role = "staff"
	HousekeepingRole Role = "housekeeping"
	FrontRole        Role = "front"
)

type FacilityHandlerInterface interface {
	GetPublicKey() *rsa.PublicKey
	GetFacilityAuthData(*gin.Context) (*FacilityAuthData, error)
	GetFacilityRole(*FacilityAuthData) (Role, error)
}

func (r Role) CanAccessStockRole() bool {
	switch r {
	case OwnerRole, ManagerRole, StaffRole:
		return true
	default:
		return false
	}
}

func isValidRole(s string) (Role, bool) {
	switch s {
	case string(OwnerRole), string(ManagerRole), string(StaffRole), string(HousekeepingRole), string(FrontRole):
		return Role(s), true
	default:
		return "", false
	}
}

func CheckFacilityToken(h FacilityHandlerInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		facilityData, err := h.GetFacilityAuthData(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "facility token is invalid", "code": "TOKEN_INVALID"})
			c.Abort()
			return
		}
		role, err := h.GetFacilityRole(facilityData)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "facility token is invalid", "code": "TOKEN_INVALID"})
			c.Abort()
			return
		}
		if !role.CanAccessStockRole() {
			c.JSON(http.StatusForbidden, gin.H{"message": "access denied", "code": "ACCESS_DENIED"})
			c.Abort()
			return
		}
		c.Set("facilityRole", role)
	}
}
