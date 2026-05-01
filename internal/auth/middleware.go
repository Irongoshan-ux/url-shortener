package auth

import (
	"context"
	"net/http"
)

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
