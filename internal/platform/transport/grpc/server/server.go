package server

import (
	"context"
	"fmt"
	"net"

	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	server *grpc.Server
	log    *logger.Logger
	config Config
}

func NewGRPCServer(cfg Config, log *logger.Logger, interceptors ...grpc.UnaryServerInterceptor) *GRPCServer {
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))

	return &GRPCServer{
		server: server,
		log:    log,
		config: cfg,
	}
}

func (s *GRPCServer) Registrar() grpc.ServiceRegistrar {
	return s.server
}

func (s *GRPCServer) Run() error {
	s.log.Warn("starting grpc server", zap.String("addr", s.config.Addr))

	l, err := net.Listen("tcp", s.config.Addr)

	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	if err := s.server.Serve(l); err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}

	return nil
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.log.Warn("shutdown grpc server", zap.String("addr", s.config.Addr))

	done := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.log.Warn("grpc server stopped", zap.String("addr", s.config.Addr))
		return nil

	case <-ctx.Done():
		s.log.Error(
			"grpc graceful shutdown timed out, forcing stop",
			zap.String("addr", s.config.Addr),
		)

		s.server.Stop()

		return fmt.Errorf("shutdown grpc server: %w", ctx.Err())
	}
}
