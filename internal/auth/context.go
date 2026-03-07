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
	ErrNoUserIDInContext   = errors.New("user id not in context")
	ErrInvalidUserIDInContext = errors.New("user id in context has invalid type")
)

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

func HadValidCookieFromContext(ctx context.Context) bool {
	v := ctx.Value(contextKeyHadValidCookie)
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func HadCookieInRequestFromContext(ctx context.Context) bool {
	v := ctx.Value(contextKeyHadCookieInRequest)
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

func WithHadValidCookie(ctx context.Context, v bool) context.Context {
	return context.WithValue(ctx, contextKeyHadValidCookie, v)
}

func WithHadCookieInRequest(ctx context.Context, v bool) context.Context {
	return context.WithValue(ctx, contextKeyHadCookieInRequest, v)
}
