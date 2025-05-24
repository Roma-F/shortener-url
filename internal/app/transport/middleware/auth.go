package middleware

import (
	"net/http"

	"github.com/Roma-F/shortener-url/internal/app/auth"
	"github.com/Roma-F/shortener-url/internal/app/logger"
)

func WithAuthentication(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := auth.GetUserIDFromCookie(r, secretKey)
			if userID == "" {
				userID = auth.GenerateUserID()
				auth.SetUserCookie(w, userID, secretKey)
				logger.Sugar.Debugw("Generated new user ID", "userID", userID)
			} else {
				logger.Sugar.Debugw("Retrieved user ID from cookie", "userID", userID)
			}

			ctx := auth.SetUserIDToContext(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
