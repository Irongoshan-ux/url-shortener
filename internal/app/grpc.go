package app

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/config"
	grpcauth "github.com/Irongoshan-ux/url-shortener/internal/grpc"
	"github.com/Irongoshan-ux/url-shortener/internal/handler"
	shortenerv1 "github.com/Irongoshan-ux/url-shortener/pkg/shortener/v1"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NewGRPCServer builds and registers the gRPC ShortenerService.
func NewGRPCServer(cfg *config.Config, facade *handler.ShortenerFacade, log zerolog.Logger) (*grpc.Server, error) {
	var opts []grpc.ServerOption
	if cfg.EnableHTTPS {
		cert, err := GenerateSelfSignedCert()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
	}
	opts = append(opts, grpc.UnaryInterceptor(grpcauth.AuthUnaryInterceptor(cfg.CookieSecret)))

	srv := grpc.NewServer(opts...)
	shortenerv1.RegisterShortenerServiceServer(srv, grpcauth.NewServer(facade, log))
	return srv, nil
}

// RunGRPC starts the gRPC server and stops it gracefully when ctx is cancelled.
func RunGRPC(ctx context.Context, cfg *config.Config, srv *grpc.Server, addr string) error {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return fmt.Errorf("grpc server address is required")
	}
	if strings.Contains(addr, "://") {
		return fmt.Errorf("grpc server address must be host:port (got %q)", addr)
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return fmt.Errorf("invalid grpc server address %q: %w", addr, err)
	}

	var ln net.Listener
	var err error
	ln, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		stopped := make(chan struct{})
		go func() {
			srv.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			return nil
		case <-time.After(shutdownTimeout):
			srv.Stop()
			return nil
		}
	}
}
