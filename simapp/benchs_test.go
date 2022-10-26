package simapp_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/simapp"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

var app *simapp.SimApp

func setup() {
	// TODO: change this
	db, err := dbm.NewGoLevelDB("application", "/Users/facundo/Downloads/testinghomedir/data")
	if err != nil {
		panic(err)
	}
	logger := log.TestingLogger()
	app = simapp.NewSimApp(logger, db, nil, true, simtestutil.NewAppOptionsWithFlagHome("/Users/facundo/Downloads/testinghomedir"))
}

func BenchmarkQueryBalance(b *testing.B) {
	b.StopTimer()
	if app == nil {
		setup()
	}
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, err := app.BankKeeper.Balance(sdk.WrapSDKContext(ctx), &banktypes.QueryBalanceRequest{Address: "cosmos1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u0tvx7u", Denom: "uatom"})
		if err != nil {
			b.Fatal(err)
		}
		if !res.Balance.Amount.Equal(sdk.NewInt(742318)) {
			b.Fatal("wrong balance", res.Balance.Amount.Int64(), 742318)
		}
	}
}

func BenchmarkQueryAllBalances(b *testing.B) {
	b.StopTimer()
	if app == nil {
		setup()
	}
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, err := app.BankKeeper.AllBalances(sdk.WrapSDKContext(ctx), &banktypes.QueryAllBalancesRequest{Address: "cosmos1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u0tvx7u"})
		if err != nil {
			b.Fatal(err)
		}

		if !res.Balances.AmountOf("uatom").Equal(sdk.NewInt(742318)) {
			b.Fatal("wrong balance", res.Balances.AmountOf("uatom").Int64(), 742318)
		}
	}
}

func BenchmarkQueryProposals(b *testing.B) {
	b.StopTimer()
	if app == nil {
		setup()
	}
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	// rpc error: code = Internal desc = no concrete type registered for type URL /cosmos.gov.v1beta1.TextProposal against interface *types.Msg
	app.InterfaceRegistry().RegisterImplementations((*sdk.Msg)(nil), &v1.MsgExecLegacyContent{})

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, err := app.GovKeeper.Proposals(sdk.WrapSDKContext(ctx), &govtypes.QueryProposalsRequest{})
		if err != nil {
			b.Fatal(err)
		}

		fmt.Printf("%#v", res)
	}
}

func BenchmarkQueryVotes(b *testing.B) {
	b.StopTimer()
	if app == nil {
		setup()
	}
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// TODO: figure out a proposal ID to query against
		res, err := app.GovKeeper.Votes(sdk.WrapSDKContext(ctx), &govtypes.QueryVotesRequest{ProposalId: 79, Pagination: &query.PageRequest{Limit: 1500}})
		if err != nil {
			b.Fatal(err)
		}

		if len(res.Votes) != 1500 {
			b.Fatal("wrong number of votes", len(res.Votes))
		}
	}
}

func BenchmarkQuerySigningInfos(b *testing.B) {
	b.StopTimer()
	if app == nil {
		setup()
	}
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, err := app.SlashingKeeper.SigningInfos(sdk.WrapSDKContext(ctx), &slashingtypes.QuerySigningInfosRequest{Pagination: &query.PageRequest{Limit: 1500}})
		if err != nil {
			b.Fatal(err)
		}

		if len(res.Info) != 402 {
			b.Fatal("wrong number of signing infos", len(res.Info))
		}
	}
}

func BenchmarkQueryValidatorDelegations(b *testing.B) {
	b.StopTimer()
	if app == nil {
		setup()
	}
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	valAddr, err := sdk.ValAddressFromBech32("cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0")
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {

		res := app.StakingKeeper.GetValidatorDelegations(ctx, valAddr)

		if len(res) != 46224 {
			b.Fatal("wrong number of delegations", len(res))
		}
	}
}
