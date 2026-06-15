package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAuthFromMetadata_NewUser(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs())
	userID, valid, hadAuth, token, err := AuthFromMetadata(ctx, "secret")
	require.NoError(t, err)
	require.NotEmpty(t, userID)
	require.False(t, valid)
	require.False(t, hadAuth)
	require.NotEmpty(t, token)
}

func TestAuthFromMetadata_ValidBearerToken(t *testing.T) {
	secret := "test-secret"
	token, err := NewToken(secret, "user-123")
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
	userID, valid, hadAuth, newToken, err := AuthFromMetadata(ctx, secret)
	require.NoError(t, err)
	require.Equal(t, "user-123", userID)
	require.True(t, valid)
	require.True(t, hadAuth)
	require.Empty(t, newToken)
}

func TestAuthFromMetadata_InvalidTokenMintsNew(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "bad-token"))
	userID, valid, hadAuth, newToken, err := AuthFromMetadata(ctx, "secret")
	require.NoError(t, err)
	require.NotEmpty(t, userID)
	require.False(t, valid)
	require.True(t, hadAuth)
	require.NotEmpty(t, newToken)
}

func TestVerifyTokenRoundTrip(t *testing.T) {
	secret := "round-trip"
	token, err := NewToken(secret, "abc")
	require.NoError(t, err)
	uid, err := VerifyToken(secret, token)
	require.NoError(t, err)
	require.Equal(t, "abc", uid)
}
