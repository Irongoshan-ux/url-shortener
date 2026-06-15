package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

// VerifyToken validates a signed JWT and returns the user id.
func VerifyToken(secret, token string) (string, error) {
	return verifyAndGetUserID(secret, token)
}

// NewToken signs a new JWT for the given user id.
func NewToken(secret, userID string) (string, error) {
	c, err := makeCookie(secret, userID)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

// AuthFromMetadata reads authorization from gRPC metadata and mirrors HTTP cookie auth semantics.
func AuthFromMetadata(ctx context.Context, secret string) (userID string, hadValidAuth, hadAuthInRequest bool, newToken string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		userID = GenerateUserID()
		newToken, err = NewToken(secret, userID)
		return userID, false, false, newToken, err
	}

	vals := md.Get("authorization")
	if len(vals) == 0 || strings.TrimSpace(vals[0]) == "" {
		userID = GenerateUserID()
		newToken, err = NewToken(secret, userID)
		return userID, false, false, newToken, err
	}

	raw := strings.TrimSpace(vals[0])
	if len(raw) > 7 && strings.EqualFold(raw[:7], "bearer ") {
		raw = strings.TrimSpace(raw[7:])
	}

	hadAuthInRequest = true
	userID, err = VerifyToken(secret, raw)
	if err != nil {
		userID = GenerateUserID()
		newToken, err = NewToken(secret, userID)
		return userID, false, true, newToken, err
	}
	return userID, true, true, "", nil
}
