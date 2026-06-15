package grpc

import (
	"context"

	"github.com/Irongoshan-ux/url-shortener/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthUnaryInterceptor validates or mints JWT from the authorization metadata header.
func AuthUnaryInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		userID, hadValid, hadAuth, newToken, err := auth.AuthFromMetadata(ctx, secret)
		if err != nil {
			return nil, err
		}
		ctx = auth.WithUserID(ctx, userID)
		ctx = auth.WithHadValidCookie(ctx, hadValid)
		ctx = auth.WithHadCookieInRequest(ctx, hadAuth)
		if newToken != "" {
			if err := grpc.SetHeader(ctx, metadata.Pairs("authorization", newToken)); err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}
