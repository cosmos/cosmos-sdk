package rosetta

import (
	"context"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	cosmos "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk"
	mocks2 "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk/mocks"
	sdktypes "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"
	mocks3 "github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint/mocks"
)

func TestLaunchpad_ConstructionMetadata(t *testing.T) {
	properties := properties{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
		AddrPrefix: "test",
	}

	networkIdentifier := types.NetworkIdentifier{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
	}

	var (
		m  = &mocks2.SdkClient{}
		mt = &mocks3.TendermintClient{}
	)

	m.
		On("GetAuthAccount", mock.Anything, "cosmos15f92rjkapauptyw6lt94rlwq4dcg99nncwc8na", int64(0)).
		Return(sdktypes.AccountResponse{
			Height: 12,
			Result: sdktypes.Response{
				Value: sdktypes.BaseAccount{
					AccountNumber: "0",
					Address:       "cosmos15f92rjkapauptyw6lt94rlwq4dcg99nncwc8na",
					Sequence:      "1",
				},
			},
		}, nil, nil).Once()

	mt.
		On("Status", mock.Anything).
		Return(tendermint.StatusResponse{
			NodeInfo: tendermint.StatusNodeInfo{
				Network: "TheNetwork",
			},
		}, nil, nil).Once()

	feeMultiplier := float64(200000)
	options := map[string]interface{}{
		OptionAddress: "cosmos15f92rjkapauptyw6lt94rlwq4dcg99nncwc8na",
		OptionGas:     &feeMultiplier,
	}

	expMetadata := map[string]interface{}{
		AccountNumberKey: "0",
		SequenceKey:      "1",
		ChainIDKey:       "TheNetwork",
		OptionGas:        &feeMultiplier,
	}

	adapter := newAdapter(nil, m, mt, properties)
	metaResp, err := adapter.ConstructionMetadata(context.Background(), &types.ConstructionMetadataRequest{
		NetworkIdentifier: &networkIdentifier,
		Options:           options,
	})

	require.Nil(t, err)
	require.NotNil(t, metaResp)
	if diff := cmp.Diff(metaResp.Metadata, expMetadata); diff != "" {
		t.Errorf("Metadata mismatch %s", diff)
	}
}

func TestLaunchpad_ConstructionMetadata_FailsOfflineMode(t *testing.T) {
	properties := properties{
		Blockchain:  "TheBlockchain",
		Network:     "TheNetwork",
		OfflineMode: true,
	}

	feeMultiplier := float64(200000)
	options := map[string]interface{}{
		OptionAddress: "cosmos15f92rjkapauptyw6lt94rlwq4dcg99nncwc8na",
		OptionGas:     &feeMultiplier,
	}

	adapter := newAdapter(nil, cosmos.NewClient("", nil), tendermint.NewClient(""), properties)
	_, err := adapter.ConstructionMetadata(context.Background(), &types.ConstructionMetadataRequest{
		Options: options,
	})

	require.Equal(t, ErrEndpointDisabledOfflineMode, err)
}
