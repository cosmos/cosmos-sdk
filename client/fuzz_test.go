package client_test

import (
	"context"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type fuzzSuite struct {
	IntegrationTestSuite
}

func (fz *fuzzSuite) FuzzQueryBalance(f *testing.F) {
	if testing.Short() {
		f.Skip("In -short mode")
	}

	// gRPC query to test service should work
	testRes, err := fz.testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	fz.Require().NoError(err)
	fz.Require().Equal("hello", testRes.Message)

	// 1. Generate some seeds.
	bz, err := fz.cdc.Marshal(&types.QueryBalanceRequest{
		Address: fz.genesisAccount.GetAddress().String(),
		Denom:   sdk.DefaultBondDenom,
	})
	fz.Require().NoError(err)
	f.Add(bz)

	// 2. Now fuzz it and ensure that we don't get any panics.
	ctx := context.Background()
	f.Fuzz(func(t *testing.T, in []byte) {
		qbReq := new(types.QueryBalanceRequest)
		if err := fz.cdc.Unmarshal(in, qbReq); err != nil {
			return
		}

		// gRPC query to bank service should work
		var header metadata.MD
		_, _ = fz.bankClient.Balance(
			ctx,
			qbReq,
			grpc.Header(&header),
		)
	})
}

func FuzzQueryBalance(f *testing.F) {
	fzs := new(fuzzSuite)
	fzs.SetT(new(testing.T))
	fzs.SetupSuite()
	fzs.FuzzQueryBalance(f)
}

func (fz *fuzzSuite) FuzzMsgClientSend(f *testing.F) {
	if testing.Short() {
		f.Skip("In -short mode")
	}

	qbL := []*types.MsgSend{
		{
			FromAddress: "cosmos1wrq8cagsama0xwf2vmlzgrkyynfsxuyhturvyz",
			ToAddress:   "cosmos1xrt7qndrz0p3kkdyvsyyjj6zwtc2ngjky8dcpe",
			Amount:      coins10,
		},
		{
			FromAddress: "cosmos1xrt7qndrz0p3kkdyvsyyjj6zwtc2ngjky8dcpe",
			ToAddress:   "cosmos1wrq8cagsama0xwf2vmlzgrkyynfsxuyhturvyz",
			Amount:      coins100,
		},
		{
			FromAddress: "cosmos1luyncewxk4lm24k6gqy8y5dxkj0klr4tu0lmnj",
			ToAddress:   "cosmos1e0jnq2sun3dzjh8p2xq95kk0expwmd7shwjpfg",
			Amount:      coins1000,
		},
	}

	for _, qb := range qbL {
		seedBlob, err := fz.cdc.Marshal(qb)
		if err != nil {
			panic(err)
		}
		f.Add(seedBlob)
	}

	ctx := context.Background()
	mc := fz.msgClient

	f.Fuzz(func(t *testing.T, inputJSONBlob []byte) {
		msg := new(types.MsgSend)
		if err := fz.cdc.Unmarshal(inputJSONBlob, msg); err != nil {
			return
		}
		if strings.TrimSpace(msg.ToAddress) == "" {
			return
		}
		if strings.TrimSpace(msg.FromAddress) == "" {
			return
		}

		_, err := mc.Send(ctx, msg)

		switch {
		case strings.Contains(err.Error(), "bech32 failed:"):
			return
		case strings.Contains(err.Error(), "invalid denom:"):
			return
		default:
			t.Fatal(err)
		}
	})
}

func FuzzMsgSend(f *testing.F) {
	fzs := new(fuzzSuite)
	fzs.SetT(new(testing.T))
	fzs.SetupSuite()
	fzs.FuzzMsgClientSend(f)
}
