package grpc

import (
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/server/config"
)

// StartGRPCWeb starts a gRPC-Web server on the given address.
func StartGRPCWeb(grpcSrv *grpc.Server, config config.Config) (*http.Server, error) {
	wrappedServer := grpcweb.WrapServer(grpcSrv)
	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedServer.ServeHTTP(resp, req)
	}
	grpcWebSrv := &http.Server{
		Addr:    config.GRPCWeb.Address,
		Handler: http.HandlerFunc(handler),
	}
	if err := grpcWebSrv.ListenAndServe(); err != nil {
		return nil, err
	}
	return grpcWebSrv, nil
}
