package rosetta

import (
	"context"
	"testing"

	cosmos "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint/mocks"

	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/require"
)

func TestLaunchpad_Mempool(t *testing.T) {
	m := &mocks.TendermintClient{}
	defer m.AssertExpectations(t)

	m.
		On("UnconfirmedTxs").
		Return(tendermint.UnconfirmedTxsResponse{
			Txs: []string{
				"1QEoKBapCl0l5qD4CiRkNGFiMDdlYi1jZGUxLTRjZmQtOWI3OS04MzYzNjFmN2RjNTcSFKeCHRQzgA2HavcLTcf4xdScUjrtGghtYW5vbGV0ZSIRdXNlckBtYW5vbGV0ZS5jb20SBBDAmgwaagom61rphyECU9fDYFDAP5TWDimv6z0BdK6oyV\nzv3iCb9fUWAAb4AoYSQCbvAfmO+aqF5WZ1M67XLZbV7OI3Sq8sbnV58tx5gf3nW/C/89pTTNmWmBskrOzmbmNEmBPQl1biuXAsUCwyMfE=",
			},
		}, nil, nil)

	adapter := newAdapter(cosmos.NewClient(""), m, properties{})

	mempool, err := adapter.Mempool(context.Background(), &types.NetworkRequest{})
	require.Nil(t, err)

	require.Equal(t, &types.MempoolResponse{TransactionIdentifiers: []*types.TransactionIdentifier{
		{
			Hash: "99b044765216517005cf096e26111016be457454ca7f83d5498d4b1142c89631",
		},
	}}, mempool)
}

func TestLaunchpad_Mempool_FailsOfflineMode(t *testing.T) {
	properties := properties{
		OfflineMode: true,
	}
	adapter := newAdapter(cosmos.NewClient(""), tendermint.NewClient(""), properties)

	_, err := adapter.Mempool(context.Background(), &types.NetworkRequest{})
	require.Equal(t, ErrEndpointDisabledOfflineMode, err)
}

func TestLaunchpad_MempoolTransaction(t *testing.T) {
	ma := &mocks.TendermintClient{}
	defer ma.AssertExpectations(t)

	ma.
		On("Tx", "ABCTHEHASH").
		Return(tendermint.TxResponse{
			Hash: "ABCTHEHASH",
		},
			nil, nil)

	adapter := newAdapter(cosmos.NewClient(""), ma, properties{})
	res, err := adapter.MempoolTransaction(context.Background(), &types.MempoolTransactionRequest{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: "ABCTHEHASH"},
	})

	require.Nil(t, err)

	require.Equal(t, "ABCTHEHASH", res.Transaction.TransactionIdentifier.Hash)
}

func TestLaunchpad_MempoolTransaction_FailsOfflineMode(t *testing.T) {
	properties := properties{
		OfflineMode: true,
	}
	adapter := newAdapter(cosmos.NewClient(""), tendermint.NewClient(""), properties)
	_, err := adapter.MempoolTransaction(context.Background(), &types.MempoolTransactionRequest{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: "ABCTHEHASH"},
	})

	require.Equal(t, ErrEndpointDisabledOfflineMode, err)
}
