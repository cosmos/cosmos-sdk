package rosetta

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"

	cosmos "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"
	types2 "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"
)

func TestPayloadsEndpoint_Errors(t *testing.T) {
	tests := []struct {
		name        string
		req         *types.ConstructionPayloadsRequest
		expectedErr *types.Error
	}{
		{
			name: "Invalid num of operations",
			req: &types.ConstructionPayloadsRequest{
				Operations: []*types.Operation{
					{
						Type: OperationTransfer,
					},
				},
			},
			expectedErr: ErrInvalidOperation,
		},
		{
			name: "Two operations not equal to transfer",
			req: &types.ConstructionPayloadsRequest{
				Operations: []*types.Operation{
					{
						Type: OperationTransfer,
					},
					{
						Type: "otherType",
					},
				},
			},
			expectedErr: rosetta.WrapError(ErrInvalidOperation, "the operations are not Transfer"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			adapter := newAdapter(nil, cosmos.NewClient("", nil), tendermint.NewClient(""), properties{})
			_, err := adapter.ConstructionPayloads(context.Background(), tt.req)
			require.Equal(t, err, tt.expectedErr)
		})
	}
}

func TestGetSenderByOperations(t *testing.T) {
	properties := properties{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
	}
	_ = newAdapter(nil, cosmos.NewClient("", nil), tendermint.NewClient(""), properties)
	ops := []*types.Operation{
		{
			Account: &types.AccountIdentifier{
				Address: "cosmos15tltvs59rt88geyenetv3klavlq2z30fe8z6hj",
			},
			Type: OperationTransfer,
			Amount: &types.Amount{
				Value: "12345",
				Currency: &types.Currency{
					Symbol:   "stake",
					Decimals: 0,
				},
				Metadata: nil,
			},
		},
		{
			Account: &types.AccountIdentifier{
				Address: "cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv",
			},
			Type: OperationTransfer,
			Amount: &types.Amount{
				Value: "-12345",
				Currency: &types.Currency{
					Symbol:   "stake",
					Decimals: 0,
				},
				Metadata: nil,
			},
		},
	}

	transferData, err := getTransferTxDataFromOperations(ops)
	require.NoError(t, err)

	expectedFrom, err := types2.AccAddressFromBech32(ops[1].Account.Address)
	require.NoError(t, err)
	expectedTo, err := types2.AccAddressFromBech32(ops[0].Account.Address)
	require.NoError(t, err)

	require.Equal(t, expectedFrom, transferData.From)
	require.Equal(t, expectedTo, transferData.To)
	require.Equal(t, types2.NewCoin("stake", types2.NewInt(12345)), transferData.Amount)
}

func TestLaunchpad_ConstructionPayloads(t *testing.T) {
	properties := properties{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
	}
	cdc, _ := simapp.MakeCodecs()
	adapter := newAdapter(cdc, cosmos.NewClient("", cdc), tendermint.NewClient(""), properties)

	feeMultiplier := float64(200000)
	senderAddr := "cosmos1khy4gsp06srvu3u65uyhrax7tnj2atez9ewh38"
	req := &types.ConstructionPayloadsRequest{
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                OperationTransfer,
				Account: &types.AccountIdentifier{
					Address: senderAddr,
				},
				Amount: &types.Amount{
					Value: "-5619726348293826415",
					Currency: &types.Currency{
						Symbol:   "atom", // TODO: Panic when bad symbol.
						Decimals: 18,
					},
				},
			},
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 1},
				RelatedOperations: []*types.OperationIdentifier{
					{Index: 0},
				},
				Type: OperationTransfer,
				Account: &types.AccountIdentifier{
					Address: "cosmos13qmcwpacu0zvsr7edpmasyn99pmcztvjhtctuz",
				},
				Amount: &types.Amount{
					Value: "5619726348293826415",
					Currency: &types.Currency{
						Symbol:   "atom",
						Decimals: 18,
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			ChainIDKey:       "theChainId",
			AccountNumberKey: "11",
			SequenceKey:      "12",
			OptionGas:        feeMultiplier,
		},
	}

	resp, err := adapter.ConstructionPayloads(context.Background(), req)
	require.Nil(t, err)

	// TODO: Decode tx and check equality
	//expectedUnssignedTx := "7b226163636f756e745f6e756d626572223a223131222c22636861696e5f6964223a22746865436861696e4964222c22666565223a7b22616d6f756e74223a5b5d2c22676173223a2230227d2c226d656d6f223a22544f444f206d656d6f222c226d736773223a5b7b2274797065223a22636f736d6f732d73646b2f4d736753656e64222c2276616c7565223a7b22616d6f756e74223a5b7b22616d6f756e74223a2235363139373236333438323933383236343135222c2264656e6f6d223a2261746f6d227d5d2c2266726f6d5f61646472657373223a22636f736d6f73316b6879346773703036737276753375363575796872617837746e6a326174657a396577683338222c22746f5f61646472657373223a22636f736d6f733133716d637770616375307a767372376564706d6173796e3939706d637a74766a68746374757a227d7d5d2c2273657175656e6365223a223132227d"

	// TODO: Create a txHex and check if same
	// Unsigned tx in hex byte representation.
	//require.Equal(t, expectedUnssignedTx, resp.UnsignedTransaction)

	require.Equal(t, senderAddr, resp.Payloads[0].AccountIdentifier.Address)
	require.Equal(t, types.SignatureType("ecdsa"), resp.Payloads[0].SignatureType)
}
