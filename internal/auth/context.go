package auth

import (
	"context"
	"errors"
)

type contextKey string

const (
	contextKeyUserID             contextKey = "user_id"
	contextKeyHadValidCookie     contextKey = "had_valid_cookie"
	contextKeyHadCookieInRequest contextKey = "had_cookie_in_request"
)

var (
	// ErrNoUserIDInContext is returned when CookieMiddleware did not run or context was replaced.
	ErrNoUserIDInContext = errors.New("user id not in context")
	// ErrInvalidUserIDInContext means the context value is not a string.
	ErrInvalidUserIDInContext = errors.New("user id in context has invalid type")
)

// UserIDFromContext returns the authenticated anonymous user id set by CookieMiddleware.
func UserIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(contextKeyUserID)
	if v == nil {
		return "", ErrNoUserIDInContext
	}
	s, ok := v.(string)
	if !ok {
		return "", ErrInvalidUserIDInContext
	}
	return s, nil
}

// HadValidCookieFromContext reports whether the request carried a signature-valid user cookie.
func HadValidCookieFromContext(ctx context.Context) bool {
	v := ctx.Value(contextKeyHadValidCookie)
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// HadCookieInRequestFromContext is true if any user_id cookie was present (valid or not).
func HadCookieInRequestFromContext(ctx context.Context) bool {
	v := ctx.Value(contextKeyHadCookieInRequest)
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// WithUserID attaches userID for tests; prefer CookieMiddleware in production.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

// WithHadValidCookie sets the valid-cookie flag for tests.
func WithHadValidCookie(ctx context.Context, v bool) context.Context {
	return context.WithValue(ctx, contextKeyHadValidCookie, v)
}

// WithHadCookieInRequest sets the had-cookie flag for tests.
func WithHadCookieInRequest(ctx context.Context, v bool) context.Context {
	return context.WithValue(ctx, contextKeyHadCookieInRequest, v)
}
