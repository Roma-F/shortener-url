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

	extractedUserID, err := VerifyAndExtractUserID(signed, testSecretKey)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)

	_, err = VerifyAndExtractUserID(signed, "wrong-key")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSignature)

	corruptedSigned := signed + "corrupted"
	_, err = VerifyAndExtractUserID(corruptedSigned, testSecretKey)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSignature)

	_, err = VerifyAndExtractUserID("no-dot-here", testSecretKey)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidFormat)

	_, err = VerifyAndExtractUserID("too.many.dots.here", testSecretKey)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidFormat)
}

func TestSetAndGetUserCookie(t *testing.T) {
	userID := "test-user-123"

	w := httptest.NewRecorder()
	SetUserCookie(w, userID, testSecretKey)

	resp := w.Result()
	defer resp.Body.Close()

	cookies := resp.Cookies()
	assert.Equal(t, 1, len(cookies))

	cookie := cookies[0]
	assert.Equal(t, UserCookieName, cookie.Name)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, CookieMaxAge, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)

	extractedUserID, err := GetUserIDFromCookie(req, testSecretKey)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}

func TestGetUserIDFromCookie_NoCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	userID, err := GetUserIDFromCookie(req, testSecretKey)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCookieNotFound)
	assert.Empty(t, userID)
}

func TestGetUserIDFromCookie_InvalidCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  UserCookieName,
		Value: "invalid-cookie-value",
	})

	userID, err := GetUserIDFromCookie(req, testSecretKey)
	assert.Error(t, err)
	assert.Empty(t, userID)
}

func TestGetUserIDFromCookie_WrongSecret(t *testing.T) {
	userID := "test-user-123"

	w := httptest.NewRecorder()
	SetUserCookie(w, userID, testSecretKey)

	resp := w.Result()
	defer resp.Body.Close()

	cookie := resp.Cookies()[0]
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)

	extractedUserID, err := GetUserIDFromCookie(req, "wrong-secret")
	assert.Error(t, err)
	assert.Empty(t, extractedUserID)
}

func TestContextOperations(t *testing.T) {
	userID := "test-user-123"
	ctx := context.Background()

	ctxWithUser, err := SetUserIDToContext(ctx, userID)
	assert.NoError(t, err)
	assert.NotEqual(t, ctx, ctxWithUser)

	extractedUserID := GetUserIDFromContext(ctxWithUser)
	assert.Equal(t, userID, extractedUserID)

	emptyUserID := GetUserIDFromContext(ctx)
	assert.Empty(t, emptyUserID)

	_, err = SetUserIDToContext(ctx, "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyUserID)

	type testKeyType string
	const testKey testKeyType = "test-key"
	ctxWithWrongType := context.WithValue(ctx, testKey, 123)
	wrongTypeUserID := GetUserIDFromContext(ctxWithWrongType)
	assert.Empty(t, wrongTypeUserID)
}
