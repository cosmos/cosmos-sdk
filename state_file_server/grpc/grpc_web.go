package grpc

import (
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/state_file_server/config"
)

// StartGRPCWeb starts a gRPC-Web server on the given address.
func StartGRPCWeb(grpcSrv *grpc.Server, config config.StateServerConfig) (*http.Server, error) {
	wrappedServer := grpcweb.WrapServer(grpcSrv)
	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedServer.ServeHTTP(resp, req)
	}
	grpcWebSrv := &http.Server{
		Addr:    config.GRPCWebAddress,
		Handler: http.HandlerFunc(handler),
	}
	if err := grpcWebSrv.ListenAndServe(); err != nil {
		return nil, err
	}
	return grpcWebSrv, nil
}
