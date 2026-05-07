// Package auth signs and verifies the user_id cookie and stores derived values in request context for handlers.
package auth

import (
	"context"
	"net/http"
)

// CookieMiddleware assigns or refreshes a JWT cookie with an anonymous user id and passes userID into request context.
func CookieMiddleware(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, hadValidCookie, hadCookieInRequest, setCookie := GetOrCreateUserID(r, secret)
			ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
			ctx = context.WithValue(ctx, contextKeyHadValidCookie, hadValidCookie)
			ctx = context.WithValue(ctx, contextKeyHadCookieInRequest, hadCookieInRequest)
			if setCookie != nil {
				http.SetCookie(w, setCookie)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
