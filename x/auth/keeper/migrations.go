package keeper

import (
	"fmt"
	"log"

	"github.com/gogo/protobuf/grpc"
	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper      AccountKeeper
	queryServer grpc.Server
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper AccountKeeper, queryServer grpc.Server) Migrator {
	return Migrator{keeper: keeper, queryServer: queryServer}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	const fqPath = "/cosmos.bank.v1beta1.Query/AllBalances"
	querier, ok := m.queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		panic(fmt.Sprintf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", m.queryServer))
	}
	log.Printf("%#v", querier)
	queryFn := querier.Route(fqPath)
	m.keeper.IterateAccounts(ctx, func(account types.AccountI) (stop bool) {
		q := &banktypes.QueryAllBalancesRequest{
			Address:    account.GetAddress().String(),
			Pagination: nil,
		}
		b, err := proto.Marshal(q)
		if err != nil {
			panic(err)
		}
		req := abci.RequestQuery{
			Data: b,
			Path: fqPath,
		}
		resp, err := queryFn(ctx, req)
		if err != nil {
			panic(err)
		}
		balance := new(banktypes.QueryAllBalancesResponse)
		err = proto.Unmarshal(resp.Value, balance)
		log.Printf("%s", balance.String())
		return false
	})

	return nil
}
