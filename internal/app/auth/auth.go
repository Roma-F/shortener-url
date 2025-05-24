package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/Roma-F/shortener-url/internal/app/logger"
)

const (
	UserCookieName = "user_id"
	CookieMaxAge   = 30 * 24 * 60 * 60
)

type UserIDKeyType string

const UserIDKey UserIDKeyType = "userID"

func GetUserIDFromCookie(r *http.Request, secretKey string) string {
	cookie, err := r.Cookie(UserCookieName)
	if err != nil {
		return ""
	}

	userID, valid := VerifyAndExtractUserID(cookie.Value, secretKey)
	if !valid {
		logger.Sugar.Debugw("Invalid cookie signature", "cookie", cookie.Value)
		return ""
	}

	return userID
}

func SetUserCookie(w http.ResponseWriter, userID, secretKey string) {
	signedUserID := SignUserID(userID, secretKey)

	cookie := &http.Cookie{
		Name:     UserCookieName,
		Value:    signedUserID,
		Path:     "/",
		MaxAge:   CookieMaxAge,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}

func GenerateUserID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		logger.Sugar.Errorw("Failed to generate random bytes for user ID", "error", err)
		return hex.EncodeToString([]byte(time.Now().String()))[:16]
	}
	return hex.EncodeToString(bytes)
}

func SignUserID(userID, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return userID + "." + signature
}

func VerifyAndExtractUserID(signedUserID, secretKey string) (string, bool) {
	parts := strings.Split(signedUserID, ".")
	if len(parts) != 2 {
		return "", false
	}

	userID := parts[0]
	expectedSigned := SignUserID(userID, secretKey)

	return userID, hmac.Equal([]byte(signedUserID), []byte(expectedSigned))
}

func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

func SetUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
