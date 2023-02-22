package grpc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/server/config"
)

// StartGRPCWeb starts a gRPC-Web server on the given address.
func StartGRPCWeb(ctx context.Context, logger log.Logger, grpcSrv *grpc.Server, config config.Config) error {
	var options []grpcweb.Option
	if config.GRPCWeb.EnableUnsafeCORS {
		options = append(options,
			grpcweb.WithOriginFunc(func(origin string) bool {
				return true
			}),
		)
	}

	wrappedServer := grpcweb.WrapServer(grpcSrv, options...)
	grpcWebSrv := &http.Server{
		Addr:              config.GRPCWeb.Address,
		Handler:           wrappedServer,
		ReadHeaderTimeout: 500 * time.Millisecond,
	}

	errCh := make(chan error)
	go func() {
		logger.Info("starting gRPC web server...", "address", config.GRPCWeb.Address)
		if err := grpcWebSrv.ListenAndServe(); err != nil {
			errCh <- fmt.Errorf("[grpc] failed to serve: %w", err)
		}
	}()

	// Start a blocking select to wait for an indication to stop the server or that
	// the server failed to start properly.
	select {
	case <-ctx.Done():
		// The calling process cancelled or closed the provided context, so we must
		// gracefully stop the gRPC-web server.
		logger.Info("stopping gRPC web server...", "address", config.GRPCWeb.Address)
		grpcWebSrv.Close()
		return nil

	case err := <-errCh:
		logger.Error("failed to start gRPC Web server", "err", err)
		return err
	}
}
