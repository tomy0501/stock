package authorize

import (
	"crypto/rsa"
	"errors"
	"log/slog"
	"net/http"
	"stock/db"
	"stock/myRedis"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type authHandler struct {
	publicKey *rsa.PublicKey
	// インターフェースから定義してモックが作れるような設計する
	db                     db.UserDB
	redis                  myRedis.RedisPool
	apiKey                 string
	apiUrl                 string
	facilityTokenPublicKey *rsa.PublicKey
}

func (a *authHandler) GetPublicKey() *rsa.PublicKey {
	return a.publicKey
}

func (a *authHandler) GetAPIKey() string {
	return a.apiKey
}

func (a *authHandler) GetServerURL() string {
	return a.apiUrl
}

func (a *authHandler) GetFacilityTokenPublicKey() *rsa.PublicKey {
	return a.facilityTokenPublicKey
}

func (a *authHandler) GetAccessToken(c *gin.Context) (*CustomClaims, error) {
	cookieString, err := c.Cookie("accessToken")
	if err != nil {
		return nil, err
	}
	/*施設サーバーに取得したアクセストークンを送信して認証し、
	さらにアクセスできる施設ID群を取得してRedisに登録する処理を実装する
	*/
	var claims CustomClaims
	token, err := jwt.ParseWithClaims(cookieString, &claims, func(tokenString *jwt.Token) (interface{}, error) {
		if _, ok := tokenString.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return a.GetPublicKey(), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}
	c.Set("accessToken", cookieString)
	return &claims, nil
}

func (h *authHandler) GetFacilityToken(c *gin.Context) (*FacilityClaims, error) {
	cookieString, err := c.Cookie("facilityToken")
	if err != nil {
		if errors.Is(http.ErrNoCookie, err) {
			apiKey := h.GetAPIKey()
			if apiKey == "" {
				err := errors.New("APIキーがありません。")
				return nil, err
			}
			apiUrl := h.GetServerURL()
			if apiUrl == "" {
				err := errors.New("URLがありません。")
				return nil, err
			}
			accessToken := c.GetString("accessToken")
			if accessToken == "" {
				return nil, errors.New("can't exist access token in session")
			}
			httpClient := &http.Client{}
			req, err := http.NewRequest("GET", apiUrl+"/validateRole", nil)
			if err != nil {
				slog.Error("HTTPクライアントの作成に失敗しました。", "error", err)
				return nil, err
			}
			req.Header.Set("Authorization", accessToken)
			resp, err := httpClient.Do(req)
			if err != nil {
				slog.Error("APIサーバーとの通信に失敗しました。")
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				slog.Error("cant not get facility token from server", "error", resp)
				return nil, errors.New("faild connect of server")
			}
			cookieString = resp.Header.Get("X-AuthorizationToken")
			if cookieString == "" {
				return nil, errors.New("can't not get cookie string from header")
			}
		} else {
			return nil, err
		}

	}

	var facilityClaims FacilityClaims

	token, err := jwt.ParseWithClaims(cookieString, &facilityClaims, func(tokenString *jwt.Token) (interface{}, error) {
		if _, ok := tokenString.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return h.GetFacilityTokenPublicKey(), nil
	})

	if err != nil || !token.Valid {
		if errors.Is(jwt.ErrTokenExpired, err) {
			apiKey := h.GetAPIKey()
			if apiKey == "" {
				err := errors.New("APIキーがありません。")
				return nil, err
			}
			apiUrl := h.GetServerURL()
			if apiUrl == "" {
				err := errors.New("URLがありません。")
				return nil, err
			}
			accessToken := c.GetString("accessToken")
			if accessToken == "" {
				return nil, errors.New("can't exist access token in session")
			}
			httpClient := &http.Client{}
			req, err := http.NewRequest("GET", apiUrl+"/validateRole", nil)
			if err != nil {
				slog.Error("HTTPクライアントの作成に失敗しました。", "error", err)
				return nil, err
			}
			req.Header.Set("Authorization", accessToken)
			resp, err := httpClient.Do(req)
			if err != nil {
				slog.Error("APIサーバーとの通信に失敗しました。")
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				slog.Error("cant not get facility token from server", "error", resp)
				return nil, errors.New("faild connect of server")
			}
			cookieString = resp.Header.Get("X-AuthorizationToken")
			if cookieString == "" {
				return nil, errors.New("can't not get cookie string from header")
			}
			var facilityClaims FacilityClaims

			token, err := jwt.ParseWithClaims(cookieString, &facilityClaims, func(tokenString *jwt.Token) (interface{}, error) {
				if _, ok := tokenString.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, jwt.ErrSignatureInvalid
				}

				return h.GetFacilityTokenPublicKey(), nil
			})
			if err != nil || !token.Valid {
				return nil, err
			}
			return &facilityClaims, nil
		} else {
			return nil, err
		}

	}
	return &facilityClaims, nil

}

func NewAuthHandler(publicKey *rsa.PublicKey, db db.UserDB, redis myRedis.RedisPool, apiKey string, url string, stockPublicKey *rsa.PublicKey) AuthHandlerInterface {
	return &authHandler{
		publicKey:              publicKey,
		db:                     db,
		redis:                  redis,
		apiKey:                 apiKey,
		apiUrl:                 url,
		facilityTokenPublicKey: stockPublicKey,
	}
}
