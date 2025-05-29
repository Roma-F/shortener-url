package middleware

import (
	"errors"
	"net/http"

	"github.com/Roma-F/shortener-url/internal/app/auth"
	"github.com/Roma-F/shortener-url/internal/app/logger"
)

func WithAuthentication(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := auth.GetUserIDFromCookie(r, secretKey)
			if errors.Is(err, auth.ErrCookieNotFound) || errors.Is(err, auth.ErrInvalidSignature) {
				userID = auth.GenerateUserID()
				auth.SetUserCookie(w, userID, secretKey)
				logger.Sugar.Debugw("Generated new user ID", "userID", userID)
			} else if err != nil {
				logger.Sugar.Errorw("Unexpected error getting user ID from cookie", "error", err)
				userID = auth.GenerateUserID()
				auth.SetUserCookie(w, userID, secretKey)
			} else {
				logger.Sugar.Debugw("Retrieved user ID from cookie", "userID", userID)
			}

			ctx, err := auth.SetUserIDToContext(r.Context(), userID)
			if err != nil {
				logger.Sugar.Errorw("Failed to set user ID to context", "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
