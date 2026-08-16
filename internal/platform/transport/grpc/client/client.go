package grpcclient

import (
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/grpc/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	conn *grpc.ClientConn
}

func NewGRPCClient(cfg Config) (*GRPCClient, error) {
	retryOpts := []retry.CallOption{
		retry.WithCodes(codes.Unavailable, codes.DeadlineExceeded, codes.Aborted),
		retry.WithMax(cfg.MaxRetries),
		retry.WithPerRetryTimeout(cfg.PerTimeout),
	}

	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
	}

	conn, err := grpc.NewClient(
		cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(interceptor.Logging(), logOpts...),
			retry.UnaryClientInterceptor(retryOpts...),
		),
		grpc.WithIdleTimeout(0),
		grpc.WithDisableServiceConfig(),
	)

	if err != nil {
		return nil, fmt.Errorf("could not create grpc client: %w", err)
	}

	conn.Connect()

	return &GRPCClient{
		conn: conn,
	}, nil
}

func (gc *GRPCClient) Conn() grpc.ClientConnInterface {
	return gc.conn
}

func (gc *GRPCClient) Close() error {
	return gc.conn.Close()
}
