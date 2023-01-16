package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/informalsystems/tm-load-test/pkg/loadtest"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/orijtech/cosmosloadtester/clients/myabciapp"
	loadtestpb "github.com/orijtech/cosmosloadtester/proto/orijtech/cosmosloadtester/v1"
	"github.com/orijtech/cosmosloadtester/server"
	"github.com/orijtech/cosmosloadtester/ui"
)

var (
	port = flag.Int("port", 8080, "the port to serve the UI and API on")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := registerClientFactories(); err != nil {
		logrus.Fatalf("failed to register client factories: %v", err)
	}

	s := server.NewServer()

	// Start the gRPC server. We don't really care what port it listens on because it will be wrapped
	// by grpc-gateway.
	lis, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		logrus.Fatalln("Failed to listen:", err)
	}
	defer lis.Close()
	grpcS := grpc.NewServer(
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)
	loadtestpb.RegisterLoadtestServiceServer(grpcS, s)
	reflection.Register(grpcS)
	logrus.Infof("Serving gRPC on %s", lis.Addr())
	go func() {
		logrus.Fatalln(grpcS.Serve(lis))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		ctx,
		lis.Addr().String(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.Fatalln("Failed to dial server:", err)
	}
	defer conn.Close()

	// Configure mux for the gRPC-Gateway API and UI.
	gwmux := runtime.NewServeMux()
	err = loadtestpb.RegisterLoadtestServiceHandler(ctx, gwmux, conn)
	if err != nil {
		logrus.Fatalln("Failed to register gateway: ", err)
	}
	fsys, err := fs.Sub(ui.UIDir, "build")
	if err != nil {
		logrus.Fatalln("failed to load embedded static content: ", err)
	}
	err = gwmux.HandlePath("GET", "/**", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.FileServer(http.FS(fsys)).ServeHTTP(w, r)
	})
	if err != nil {
		logrus.Fatalln("Failed to register static content with gateway: ", err)
	}
	wrappedGrpc := grpcweb.WrapServer(grpcS)
	wrappedHandler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(req) {
			wrappedGrpc.ServeHTTP(res, req)
			return
		}
		gwmux.ServeHTTP(res, req)
	})

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: wrappedHandler,
	}
	logrus.Infof("Serving gRPC-Gateway on http://%s", gwServer.Addr)
	go func() {
		logrus.Fatalln(gwServer.ListenAndServe())
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
}

// Add logic to register your custom client factories to this function.
func registerClientFactories() error {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	cosmosClientFactory := myabciapp.NewCosmosClientFactory(txConfig)
	if err := loadtest.RegisterClientFactory("test-cosmos-client-factory", cosmosClientFactory); err != nil {
		return fmt.Errorf("failed to register client factory %s: %w", "test-cosmos-client-factory", err)
	}
	return nil
}
