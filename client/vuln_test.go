package client_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestNodeCrashPanics(t *testing.T) {
	t.Skip("Need a running node")

	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	tc := types.NewQueryClient(conn)
	_, err = tc.Balance(
		context.Background(),
		&types.QueryBalanceRequest{
			Address: "cosmos1wrq8cagsama0xwf2vmlzgrkyynfsxuyhturvyz",
			Denom:   "00000",
		},
	)
	if err != nil {
		panic(err)
	}
}

func FuzzNodeQueryBalance(f *testing.F) {
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	tc := types.NewQueryClient(conn)

	qbL := []*types.QueryBalanceRequest{
		{
			Address: "cosmos1wrq8cagsama0xwf2vmlzgrkyynfsxuyhturvyz",
			Denom:   "stake",
		},
	}

	for _, qb := range qbL {
		seedBlob, err := json.Marshal(qb)
		if err != nil {
			panic(err)
		}
		f.Add(seedBlob)
	}

	ctx := context.Background()

	f.Fuzz(func(t *testing.T, inputJSONBlob []byte) {
		qb := new(types.QueryBalanceRequest)
		if err := json.Unmarshal(inputJSONBlob, qb); err != nil {
			return
		}
		if strings.TrimSpace(qb.Address) == "" {
			return
		}
		if strings.TrimSpace(qb.Denom) == "" {
			return
		}
		_, err := tc.Balance(ctx, qb)
		if err == nil {
			return
		}

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

var (
	tDenom    = sdk.DefaultBondDenom
	coins10   = sdk.NewCoins(sdk.NewInt64Coin(tDenom, 10))
	coins100  = sdk.NewCoins(sdk.NewInt64Coin(tDenom, 100))
	coins1000 = sdk.NewCoins(sdk.NewInt64Coin(tDenom, 1000))
)

func FuzzMsgClientSend(f *testing.F) {
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	tc := types.NewMsgClient(conn)

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
		seedBlob, err := json.Marshal(qb)
		if err != nil {
			panic(err)
		}
		f.Add(seedBlob)
	}

	ctx := context.Background()

	f.Fuzz(func(t *testing.T, inputJSONBlob []byte) {
		msg := new(types.MsgSend)
		if err := json.Unmarshal(inputJSONBlob, msg); err != nil {
			return
		}
		if strings.TrimSpace(msg.ToAddress) == "" {
			return
		}
		if strings.TrimSpace(msg.FromAddress) == "" {
			return
		}
		_, err := tc.Send(ctx, msg)
		if err == nil {
			return
		}

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
