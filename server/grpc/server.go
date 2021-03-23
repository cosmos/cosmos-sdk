package grpc

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/server/grpc/cosmosreflection"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/grpc/gogoreflection"
	"github.com/cosmos/cosmos-sdk/server/types"
)

// StartGRPCServer starts a gRPC server on the given address.
func StartGRPCServer(clientCtx client.Context, app types.Application, address string) (*grpc.Server, error) {
	grpcSrv := grpc.NewServer()
	app.RegisterGRPCServer(clientCtx, grpcSrv)
	// cosmosreflection allows consumers to build dynamic clients that can write
	// to any cosmos-sdk application without relying on application packages at compile time
	err := cosmosreflection.Register(grpcSrv, cosmosreflection.Config{
		SigningModes: func() []string {
			modes := make([]string, len(clientCtx.TxConfig.SignModeHandler().Modes()))
			for i, m := range clientCtx.TxConfig.SignModeHandler().Modes() {
				modes[i] = m.String()
			}
			return modes
		}(),
		ChainID:           clientCtx.ChainID,
		SdkConfig:         sdk.GetConfig(),
		InterfaceRegistry: clientCtx.InterfaceRegistry,
	})
	if err != nil {
		return nil, err
	}
	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	gogoreflection.Register(grpcSrv)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	errCh := make(chan error)
	go func() {
		err = grpcSrv.Serve(listener)
		if err != nil {
			errCh <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return nil, err
	case <-time.After(5 * time.Second): // assume server started successfully
		return grpcSrv, nil
	}
}
