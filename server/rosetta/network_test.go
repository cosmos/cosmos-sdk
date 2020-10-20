package rosetta

import (
	"context"
	"testing"
	"time"

	mocks2 "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk/mocks"
	mocks3 "github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint/mocks"

	cosmos "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk"

	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/tendermint/tendermint/p2p"

	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLaunchpad_NetworkList(t *testing.T) {
	properties := properties{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
	}

	adapter := newAdapter(cosmos.NewClient(""), tendermint.NewClient(""), properties)

	list, err := adapter.NetworkList(context.Background(), nil)
	require.Nil(t, err)

	require.Len(t, list.NetworkIdentifiers, 1)
	require.Equal(t, list.NetworkIdentifiers[0].Network, "TheNetwork")
	require.Equal(t, list.NetworkIdentifiers[0].Blockchain, "TheBlockchain")
}

func TestLaunchpad_NetworkList_FailsOfflineMode(t *testing.T) {
	properties := properties{
		Blockchain:  "TheBlockchain",
		Network:     "TheNetwork",
		OfflineMode: true,
	}

	adapter := newAdapter(cosmos.NewClient(""), tendermint.NewClient(""), properties)

	_, err := adapter.NetworkList(context.Background(), nil)
	require.Equal(t, err, ErrEndpointDisabledOfflineMode)
}

func TestLaunchpad_NetworkOptions(t *testing.T) {
	m := &mocks2.SdkClient{}
	defer m.AssertExpectations(t)

	m.
		On("GetNodeInfo", mock.Anything).
		Return(rpc.NodeInfoResponse{
			DefaultNodeInfo: p2p.DefaultNodeInfo{
				Version: "5",
			},
		}, nil, nil).
		Once()

	properties := properties{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
	}

	adapter := newAdapter(m, tendermint.NewClient(""), properties)

	options, err := adapter.NetworkOptions(context.Background(), nil)
	require.Nil(t, err)
	require.NotNil(t, options)

	require.Equal(t, &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion: "1.2.5",
			NodeVersion:    "5",
		},
		Allow: &types.Allow{
			OperationStatuses: []*types.OperationStatus{
				{
					Status:     StatusSuccess,
					Successful: true,
				},
				{
					Status:     StatusReverted,
					Successful: false,
				},
			},
			OperationTypes: []string{
				OperationTransfer,
			},
		},
	}, options)
}

func TestLaunchpad_NetworkOptions_FailsOfflineMode(t *testing.T) {
	properties := properties{
		Blockchain:  "TheBlockchain",
		Network:     "TheNetwork",
		OfflineMode: true,
	}

	adapter := newAdapter(cosmos.NewClient(""), tendermint.NewClient(""), properties)

	_, err := adapter.NetworkOptions(context.Background(), nil)
	require.Equal(t, err, ErrEndpointDisabledOfflineMode)
}

func TestLaunchpad_NetworkStatus(t *testing.T) {
	mt := mocks3.TendermintClient{}
	defer mt.AssertExpectations(t)

	ti, err := time.Parse(time.RFC3339, "2019-04-22T17:01:51Z")
	require.NoError(t, err)

	mt.
		On("Block", uint64(0)).
		Return(tendermint.BlockResponse{
			Block: tendermint.Block{
				Header: tendermint.BlockHeader{
					Time:   ti.Format(time.RFC3339),
					Height: "2",
				},
			},
			BlockID: tendermint.BlockID{
				Hash: "3",
			},
		}, nil, nil).
		Once()

	mt.
		On("Block", uint64(1)).
		Return(tendermint.BlockResponse{
			Block: tendermint.Block{
				Header: tendermint.BlockHeader{
					Height: "1",
				},
			},
			BlockID: tendermint.BlockID{
				Hash: "4",
			},
		}, nil, nil).
		Once()

	mt.
		On("NetInfo", mock.Anything).
		Return(tendermint.NetInfoResponse{
			Peers: []tendermint.Peer{
				{
					NodeInfo: tendermint.NodeInfo{
						ID: "1",
					},
				},
				{
					NodeInfo: tendermint.NodeInfo{
						ID: "2",
					},
				},
			},
		}, nil, nil).
		Once()

	properties := properties{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
	}

	adapter := newAdapter(
		cosmos.NewClient(""),
		&mt,
		properties,
	)

	status, adapterErr := adapter.NetworkStatus(context.Background(), nil)
	require.Nil(t, adapterErr)
	require.NotNil(t, status)

	require.Equal(t, &types.NetworkStatusResponse{
		CurrentBlockIdentifier: &types.BlockIdentifier{
			Index: 2,
			Hash:  "3",
		},
		CurrentBlockTimestamp: ti.UnixNano() / 1000000,
		GenesisBlockIdentifier: &types.BlockIdentifier{
			Index: 1,
			Hash:  "4",
		},
		Peers: []*types.Peer{
			{
				PeerID: "1",
			},
			{
				PeerID: "2",
			},
		},
	}, status)
}

func TestLaunchpad_NetworkStatus_FailsOfflineMode(t *testing.T) {
	properties := properties{
		Blockchain:  "TheBlockchain",
		Network:     "TheNetwork",
		OfflineMode: true,
	}

	adapter := newAdapter(
		cosmos.NewClient(""),
		tendermint.NewClient(""),
		properties,
	)

	_, adapterErr := adapter.NetworkStatus(context.Background(), nil)
	require.Equal(t, ErrEndpointDisabledOfflineMode, adapterErr)

}
