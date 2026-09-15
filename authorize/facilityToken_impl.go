package authorize

import (
	"crypto/rsa"
	"log/slog"
	"net/http"
	"stock/myRedis"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type facilityTokenHandler struct {
	publicKey *rsa.PublicKey
	redis     myRedis.RedisPool
	apiKey    string
	url       string
}

func NewFacilityTokenHandler(publicKey *rsa.PublicKey, redis myRedis.RedisPool, apiKey string, url string) FacilityHandlerInterface {
	return &facilityTokenHandler{
		publicKey: publicKey,
		redis:     redis,
		apiKey:    apiKey,
		url:       url,
	}
}

func (f *facilityTokenHandler) GetPublicKey() *rsa.PublicKey {
	return f.publicKey
}

func (f *facilityTokenHandler) GetFacilityAuthData(c *gin.Context) (*FacilityAuthData, error) {
	// facilityTokenを取得する処理を実装する
	// ここでは、facilityTokenを取得するための処理を実装する必要があります。
	// 例えば、リクエストヘッダーからfacilityTokenを取得し、デコードしてFacilityAuthDataを返すなどの処理が考えられます。
	// ここでは厳密なユーザーIDの検証は行わず他のサーバーにユーザーIDと施設IDを渡して検証する想定です。
	facilityCookie, err := c.Cookie("facilityToken")
	if err != nil {
		return nil, err
	}
	userIDCookie, err := c.Cookie("userID")
	if err != nil {
		return nil, err
	}
	facilityData := &FacilityAuthData{
		FacilityID: facilityCookie,
		UserID:     userIDCookie,
	}
	return facilityData, nil
}

func (f *facilityTokenHandler) GetFacilityRole(facilityData *FacilityAuthData) (Role, error) {
	// facilityDataからRoleを取得する処理を実装する
	// ここでは、facilityDataからRoleを取得するための処理を実装する必要があります。
	// 例えば、facilityDataのUserIDを使ってデータベースからRoleを取得するなどの処理が考えられます。
	// ここでは別のサーバーにユーザーIDと施設IDを渡して整合性とRoleを検証する想定です。
	checkRedisData, err := f.redis.Load(facilityData.FacilityID)
	if err == nil {
		isExistUserID, ok := checkRedisData[facilityData.UserID]
		if ok {
			role, isRole := isValidRole(isExistUserID.(string))
			if isRole {
				return role, nil
			}
			slog.Warn("redisに保存されているRoleが不正です。再取得を実行します。", "facilityID", facilityData.FacilityID, "userID", facilityData.UserID)
		}
		slog.Warn("該当ユーザーのIDがredisに存在しません。再取得を実行します。", "facilityID", facilityData.FacilityID, "userID", facilityData.UserID)
	}
	apiKey := f.apiKey
	if apiKey == "" {
		slog.Error("apiキーが見つかりません。")
		return "", nil
	}
	apiUrl := f.url
	if apiUrl == "" {
		slog.Error("URLが見つかりません。")
		return "", nil
	}
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl+"/validateRole", nil)
	if err != nil {
		slog.Error("HTTPクライアントの作成に失敗しました。", "error", err)
		return "", err
	}
	req.Header.Set("X-API-Key", apiKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Error("APIサーバーとの通信に失敗しました。")
		return "", err
	}
	defer resp.Body.Close()
	// 応答を処理してRoleを取得する
	if resp.StatusCode != http.StatusOK {
		return "", http.ErrNoLocation
	}
	facilityAPIJWT := resp.Header.Get("X-Facility-Role")
	if facilityAPIJWT == "" {
		return "", http.ErrNoLocation
	}

	var facilityClaims FacilityClaim

	facilityToken, err := jwt.ParseWithClaims(facilityAPIJWT, &facilityClaims, func(tokenString *jwt.Token) (interface{}, error) {
		if _, ok := tokenString.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return f.publicKey, nil
	})
	if err != nil {
		return "", err
	}
	if !facilityToken.Valid {
		return "", jwt.ErrSignatureInvalid
	}
	facilityRole, isRole := isValidRole(facilityClaims.Role)
	if !isRole {
		return "", jwt.ErrSignatureInvalid
	}
	ttl := facilityClaims.ExpiresAt.Time.Sub(facilityClaims.IssuedAt.Time).Seconds()
	if ttl <= 0 {
		return "", jwt.ErrTokenExpired
	}

	saveWithRedis := map[string]any{
		facilityData.UserID: facilityRole,
	}
	err = f.redis.SaveWithTTL(facilityData.FacilityID, saveWithRedis, int(ttl))
	if err != nil {
		return "", err
	}
	return facilityRole, nil
}
