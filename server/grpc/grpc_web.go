package grpc

import (
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/server/config"
)

// StartGRPCWeb starts a gRPC-Web server on the given address.
func StartGRPCWeb(grpcSrv *grpc.Server, config config.Config) (*http.Server, error) {
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
		Addr:    config.GRPCWeb.Address,
		Handler: wrappedServer,
	}
	if err := grpcWebSrv.ListenAndServe(); err != nil {
		return nil, err
	}
	return grpcWebSrv, nil
}
