package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/math"
	minttypes "cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestAccountRetriever(t *testing.T) {
	t.Skip() // TODO: https://github.com/cosmos/cosmos-sdk/issues/22825

	f := initFixture(t, nil)

	grpcSrv := grpc.NewServer(grpc.ForceServerCodec(codec.NewProtoCodec(f.encodingCfg.InterfaceRegistry).GRPCCodec()))

	types.RegisterQueryServer(f.app.GRPCQueryRouter(), keeper.NewQueryServer(f.authKeeper))
	f.app.RegisterGRPCServer(grpcSrv)

	grpcCfg := srvconfig.DefaultConfig().GRPC

	go func() {
		require.NoError(t, servergrpc.StartGRPCServer(context.Background(), f.app.Logger(), grpcCfg, grpcSrv))
	}()

	conn, err := grpc.NewClient(
		grpcCfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(f.encodingCfg.InterfaceRegistry).GRPCCodec())),
	)
	require.NoError(t, err)

	defer conn.Close()

	pubkeys := simtestutil.CreateTestPubKeys(1)
	addr := sdk.AccAddress(pubkeys[0].Address())

	newAcc := types.BaseAccount{
		Address:       addr.String(),
		PubKey:        nil,
		AccountNumber: 2,
		Sequence:      7,
	}

	updatedAcc := f.authKeeper.NewAccount(f.ctx, &newAcc)
	f.authKeeper.SetAccount(f.ctx, updatedAcc)

	amount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10000)))
	require.NoError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, amount))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, minttypes.ModuleName, addr, amount))

	ar := types.AccountRetriever{}

	clientCtx := client.Context{}.
		WithGRPCClient(conn).
		WithAddressPrefix(sdk.Bech32MainPrefix)

	acc, err := ar.GetAccount(clientCtx, addr)
	require.NoError(t, err)
	require.NotNil(t, acc)

	acc, height, err := ar.GetAccountWithHeight(clientCtx, addr)
	require.NoError(t, err)
	require.NotNil(t, acc)
	require.Equal(t, height, int64(2))

	require.NoError(t, ar.EnsureExists(clientCtx, addr))

	accNum, accSeq, err := ar.GetAccountNumberSequence(clientCtx, addr)
	require.NoError(t, err)
	require.Equal(t, accNum, uint64(0))
	require.Equal(t, accSeq, uint64(1))
}
