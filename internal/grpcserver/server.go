package grpcserver

import (
	"context"
	"errors"

	"github.com/Irongoshan-ux/url-shortener/internal/auth"
	"github.com/Irongoshan-ux/url-shortener/internal/handler"
	shortenerv1 "github.com/Irongoshan-ux/url-shortener/pkg/shortener/v1"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server implements shortenerv1.ShortenerServiceServer.
type Server struct {
	shortenerv1.UnimplementedShortenerServiceServer
	facade *handler.ShortenerFacade
	log    zerolog.Logger
}

// NewServer constructs a gRPC ShortenerService implementation.
func NewServer(facade *handler.ShortenerFacade, log zerolog.Logger) *Server {
	return &Server{
		facade: facade,
		log:    log,
	}
}

// ShortenURL creates a shortened URL for the authenticated user.
func (s *Server) ShortenURL(ctx context.Context, req *shortenerv1.URLShortenRequest) (*shortenerv1.URLShortenResponse, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("user id not in context")
		return nil, status.Error(codes.Internal, "internal error")
	}
	shortID, _, err := s.facade.ShortenURL(ctx, req.GetUrl(), userID)
	if err != nil {
		if errors.Is(err, handler.ErrEmptyURL) || errors.Is(err, handler.ErrInvalidURL) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Info().Err(err).Msg("shorten url failed")
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &shortenerv1.URLShortenResponse{Result: s.facade.BuildFullURLFromBase(shortID)}, nil
}

// ExpandURL resolves a short id to the original URL.
func (s *Server) ExpandURL(ctx context.Context, req *shortenerv1.URLExpandRequest) (*shortenerv1.URLExpandResponse, error) {
	originalURL, err := s.facade.ExpandURL(ctx, req.GetId())
	if err != nil {
		switch {
		case errors.Is(err, handler.ErrExpandNotFound), errors.Is(err, handler.ErrExpandGone):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			s.log.Info().Err(err).Msg("expand url failed")
			return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return &shortenerv1.URLExpandResponse{Result: originalURL}, nil
}

// ListUserURLs returns URLs owned by the authenticated user.
func (s *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*shortenerv1.UserURLsResponse, error) {
	if !auth.HadValidCookieFromContext(ctx) {
		if auth.HadCookieInRequestFromContext(ctx) {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}
		return &shortenerv1.UserURLsResponse{}, nil
	}
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		s.log.Error().Err(err).Msg("user id not in context")
		return nil, status.Error(codes.Internal, "internal error")
	}
	items, err := s.facade.ListUserURLs(ctx, userID, nil)
	if err != nil {
		s.log.Info().Err(err).Msg("list user urls failed")
		return nil, status.Error(codes.Internal, "internal error")
	}
	if len(items) == 0 {
		return &shortenerv1.UserURLsResponse{}, nil
	}
	urls := make([]*shortenerv1.URLData, len(items))
	for i, item := range items {
		urls[i] = &shortenerv1.URLData{
			ShortUrl:    item.ShortURL,
			OriginalUrl: item.OriginalURL,
		}
	}
	return &shortenerv1.UserURLsResponse{Url: urls}, nil
}
