package rosetta

import (
	"context"
	"testing"

	cosmos "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestLaunchpad_ConstructionPreprocess(t *testing.T) {
	properties := properties{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
		AddrPrefix: "test",
	}
	adapter := newAdapter(nil, cosmos.NewClient("", nil), tendermint.NewClient(""), properties)

	ops := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{},
			Type:                "Transfer",
			Status:              "Success",
			Account: &types.AccountIdentifier{
				Address: "cosmos12qqzw4tqu32anlcx0a3hupvgdhaf4cc8j9wfyd",
			},
			Amount: &types.Amount{
				Value: "-10",
				Currency: &types.Currency{
					Symbol: "token",
				},
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 1,
			},
			Type:   "Transfer",
			Status: "Success",
			Account: &types.AccountIdentifier{
				Address: "cosmos10rpmm9ur87le39hehteha37sg5awdsns6huyvy",
			},
			Amount: &types.Amount{
				Value: "10",
				Currency: &types.Currency{
					Symbol: "token",
				},
			},
		},
	}
	feeMultiplier := float64(200000)

	expOptions := map[string]interface{}{
		OptionAddress: "cosmos12qqzw4tqu32anlcx0a3hupvgdhaf4cc8j9wfyd",
		OptionGas:     200000,
	}

	deriveResp, deriveErr := adapter.ConstructionPreprocess(context.Background(), &types.ConstructionPreprocessRequest{
		Operations:             ops,
		SuggestedFeeMultiplier: &feeMultiplier,
	})

	require.Nil(t, deriveErr)
	require.NotNil(t, deriveResp)
	if diff := cmp.Diff(deriveResp.Options, expOptions); diff != "" {
		t.Errorf("properties mismatch %s", diff)
	}
}
