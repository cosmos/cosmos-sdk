package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func queryState(t *testing.T) {
	myAddress, err := sdk.AccAddressFromBech32("cosmos1v0htrgupr0ans0zu2a94l6w9w6n82t49333g8r") // the my_validator or recipient address.
	require.NoError(t, err)

	appConfig := network.MinimumAppConfig()
	var (
		txConfig          client.TxConfig
		legacyAmino       *codec.LegacyAmino
		cdc               codec.Codec
		interfaceRegistry codectypes.InterfaceRegistry
	)

	err = depinject.Inject(
		depinject.Configs(
			appConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&txConfig,
		&cdc,
		&legacyAmino,
		&interfaceRegistry,
	)
	require.NoError(t, err)

	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		"0.0.0.0:9091",      // your gRPC server address.
		grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
		// This instantiates a general gRPC codec which handles proto bytes. We pass in a nil interface registry
		// if the request/response types contain interface instead of 'nil' you should pass the application specific codec.
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(interfaceRegistry).GRPCCodec())),
	)
	require.NoError(t, err)
	defer grpcConn.Close()

	rc := grpcreflect.NewClientAuto(context.Background(), grpcConn)

	services, err := rc.ListServices()
	require.NoError(t, err)
	fmt.Printf("services found: %v\n", services)

	// This creates a gRPC client to query the x/bank service.
	bankClient := banktypes.NewQueryClient(grpcConn)
	bankRes, err := bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: myAddress.String(), Denom: "stake"},
	)
	require.NoError(t, err)

	fmt.Println(bankRes.GetBalance()) // Prints the account balance
}

func TestTest(t *testing.T) {
	queryState(t)
}
