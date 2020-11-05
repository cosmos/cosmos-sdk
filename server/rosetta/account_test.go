package rosetta

import (
	"context"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/require"

	cosmos "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk"
	sdkmocks "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk/mocks"
	sdktypes "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"
	tendermintmocks "github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint/mocks"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestLaunchpad_AccountBalance(t *testing.T) {
	m := &sdkmocks.SdkClient{}
	ma := &tendermintmocks.TendermintClient{}
	defer m.AssertExpectations(t)
	defer ma.AssertExpectations(t)

	m.
		On("GetAuthAccount", context.Background(), "cosmos15f92rjkapauptyw6lt94rlwq4dcg99nncwc8na", int64(0)).
		Return(sdktypes.AccountResponse{
			Height: 12,
			Result: sdktypes.Response{
				Value: sdktypes.BaseAccount{
					AccountNumber: "0",
					Coins: []sdk.Coin{
						{Denom: "stake", Amount: sdk.NewInt(400)},
						{Denom: "token", Amount: sdk.NewInt(600)},
					},
					Address:  "cosmos15f92rjkapauptyw6lt94rlwq4dcg99nncwc8na",
					Sequence: "1",
				},
			},
		}, nil, nil).Once()

	blockHash := "ABCDEFG"
	ma.
		On("Block", uint64(12)).
		Return(tendermint.BlockResponse{
			BlockID: tendermint.BlockID{
				Hash: blockHash,
			},
			Block: tendermint.Block{},
		}, nil, nil)

	properties := properties{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
	}

	adapter := newAdapter(nil, m, ma, properties)

	res, err := adapter.AccountBalance(context.Background(), &types.AccountBalanceRequest{
		AccountIdentifier: &types.AccountIdentifier{
			Address: "cosmos15f92rjkapauptyw6lt94rlwq4dcg99nncwc8na",
		},
	})
	require.Nil(t, err)
	require.Len(t, res.Balances, 2)
	require.Equal(t, res.BlockIdentifier.Hash, blockHash)
	require.Equal(t, res.BlockIdentifier.Index, int64(12))

	// NewCoins sorts the coins by name.
	require.Equal(t, "400", res.Balances[0].Value)
	require.Equal(t, "stake", res.Balances[0].Currency.Symbol)
	require.Equal(t, "600", res.Balances[1].Value)
	require.Equal(t, "token", res.Balances[1].Currency.Symbol)
}

func TestLaunchpad_AccountBalanceDoesNotWorkOfflineMode(t *testing.T) {
	properties := properties{
		Blockchain:  "TheBlockchain",
		Network:     "TheNetwork",
		OfflineMode: true,
	}

	adapter := newAdapter(nil, cosmos.NewClient("", nil), tendermint.NewClient(""), properties)
	_, err := adapter.AccountBalance(context.Background(), &types.AccountBalanceRequest{
		AccountIdentifier: &types.AccountIdentifier{
			Address: "cosmos15f92rjkapauptyw6lt94rlwq4dcg99nncwc8na",
		},
	})

	require.Equal(t, err, ErrEndpointDisabledOfflineMode)
}
