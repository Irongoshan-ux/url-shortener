package auth

import (
	"context"
	"net/http"
)

func CookieMiddleware(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, hadValidCookie, setCookie := GetOrCreateUserID(r, secret)
			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			ctx = context.WithValue(ctx, ContextKeyHadValidCookie, hadValidCookie)
			if setCookie != nil {
				http.SetCookie(w, setCookie)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ContextKeyUserID)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func HadValidCookieFromContext(ctx context.Context) bool {
	v := ctx.Value(ContextKeyHadValidCookie)
	if v == nil {
		return false
	}
	b, _ := v.(bool)
	return b
}
