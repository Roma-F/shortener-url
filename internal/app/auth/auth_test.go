package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testSecretKey = "test-secret-key-for-testing"

func TestGenerateUserID(t *testing.T) {
	userID1 := GenerateUserID()
	userID2 := GenerateUserID()

	assert.NotEmpty(t, userID1)
	assert.NotEmpty(t, userID2)

	assert.NotEqual(t, userID1, userID2)

	assert.Equal(t, 32, len(userID1))
	assert.Equal(t, 32, len(userID2))
}

func TestSignUserID(t *testing.T) {
	userID := "test-user-123"

	signed := SignUserID(userID, testSecretKey)

	parts := strings.Split(signed, ".")
	assert.Equal(t, 2, len(parts), "Signed userID should have format 'userID.signature'")
	assert.Equal(t, userID, parts[0])
	assert.NotEmpty(t, parts[1])

	signed2 := SignUserID(userID, testSecretKey)
	assert.Equal(t, signed, signed2)

	signed3 := SignUserID(userID, "different-key")
	assert.NotEqual(t, signed, signed3)
}

func TestVerifyAndExtractUserID(t *testing.T) {
	userID := "test-user-123"
	signed := SignUserID(userID, testSecretKey)

	extractedUserID, valid := VerifyAndExtractUserID(signed, testSecretKey)
	assert.True(t, valid)
	assert.Equal(t, userID, extractedUserID)

	_, valid = VerifyAndExtractUserID(signed, "wrong-key")
	assert.False(t, valid)

	corruptedSigned := signed + "corrupted"
	_, valid = VerifyAndExtractUserID(corruptedSigned, testSecretKey)
	assert.False(t, valid)

	_, valid = VerifyAndExtractUserID("no-dot-here", testSecretKey)
	assert.False(t, valid)

	_, valid = VerifyAndExtractUserID("too.many.dots.here", testSecretKey)
	assert.False(t, valid)
}

func TestSetAndGetUserCookie(t *testing.T) {
	userID := "test-user-123"

	w := httptest.NewRecorder()
	SetUserCookie(w, userID, testSecretKey)

	cookies := w.Result().Cookies()
	assert.Equal(t, 1, len(cookies))

	cookie := cookies[0]
	assert.Equal(t, UserCookieName, cookie.Name)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, CookieMaxAge, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)

	extractedUserID := GetUserIDFromCookie(req, testSecretKey)
	assert.Equal(t, userID, extractedUserID)
}

func TestGetUserIDFromCookie_NoCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	userID := GetUserIDFromCookie(req, testSecretKey)
	assert.Empty(t, userID)
}

func TestGetUserIDFromCookie_InvalidCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  UserCookieName,
		Value: "invalid-cookie-value",
	})

	userID := GetUserIDFromCookie(req, testSecretKey)
	assert.Empty(t, userID)
}

func TestGetUserIDFromCookie_WrongSecret(t *testing.T) {
	userID := "test-user-123"

	w := httptest.NewRecorder()
	SetUserCookie(w, userID, testSecretKey)
	cookie := w.Result().Cookies()[0]

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)

	extractedUserID := GetUserIDFromCookie(req, "wrong-secret")
	assert.Empty(t, extractedUserID)
}

func TestContextOperations(t *testing.T) {
	userID := "test-user-123"
	ctx := context.Background()

	ctxWithUser := SetUserIDToContext(ctx, userID)
	assert.NotEqual(t, ctx, ctxWithUser)

	extractedUserID := GetUserIDFromContext(ctxWithUser)
	assert.Equal(t, userID, extractedUserID)

	emptyUserID := GetUserIDFromContext(ctx)
	assert.Empty(t, emptyUserID)

	ctxWithWrongType := context.WithValue(ctx, UserIDKey, 123)
	wrongTypeUserID := GetUserIDFromContext(ctxWithWrongType)
	assert.Empty(t, wrongTypeUserID)
}
