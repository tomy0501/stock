package authorize

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func CheckAccessToken(h AuthHandlerInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := h.GetAccessToken(c)
		if err != nil {
			if errors.Is(jwt.ErrTokenExpired, err) {
				slog.Info("jwt is expired please rerresh", "info", accessToken)
				c.JSON(http.StatusUnauthorized, gin.H{"message": "please token refresh", "code": "TOKEN_EXPIRED"})
				c.Abort()
				return
			} else {
				slog.Error("happen invalid accessToken error", "error", err)
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorize", "code": "INVALID_TOKEN"})
				c.Abort()
				return
			}
		}
		c.Set("accessToken", accessToken)
		userID, err := accessToken.GetSubject()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error happen"})
			c.Abort()
			return
		}
		c.Set("userID", userID)
		facilityToken, err := h.GetFacilityToken(c)
		if err != nil {
			slog.Error("when get facilityToken, happen a unknown error", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error happen"})
			c.Abort()
			return
		}
		accessFacilities := facilityToken.UserFacilitys
		c.Set("facilities", accessFacilities)

	}
}
