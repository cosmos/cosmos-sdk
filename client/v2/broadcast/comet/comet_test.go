package comet

import (
	"context"
	"errors"
	"testing"

	"github.com/cometbft/cometbft/mempool"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	apiacbci "cosmossdk.io/api/cosmos/base/abci/v1beta1"
	mockrpc "cosmossdk.io/client/v2/broadcast/comet/testutil"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
)

var cdc = testutil.CodecOptions{}.NewCodec()

func TestNewCometBftBroadcaster(t *testing.T) {
	tests := []struct {
		name    string
		cdc     codec.Codec
		mode    string
		want    *CometBFTBroadcaster
		wantErr bool
	}{
		{
			name: "constructor",
			mode: BroadcastSync,
			cdc:  cdc,
			want: &CometBFTBroadcaster{
				mode: BroadcastSync,
				cdc:  cdc,
			},
		},
		{
			name:    "nil codec",
			mode:    BroadcastSync,
			cdc:     nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCometBFTBroadcaster("localhost:26657", tt.mode, tt.cdc)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.Equal(t, got.mode, tt.want.mode)
				require.Equal(t, got.cdc, tt.want.cdc)
			}
		})
	}
}

func TestCometBftBroadcaster_Broadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	cometMock := mockrpc.NewMockCometRPC(ctrl)
	c := CometBFTBroadcaster{
		rpcClient: cometMock,
		mode:      BroadcastSync,
		cdc:       cdc,
	}
	tests := []struct {
		name      string
		mode      string
		setupMock func(*mockrpc.MockCometRPC)
		wantErr   bool
	}{
		{
			name: "sync",
			mode: BroadcastSync,
			setupMock: func(m *mockrpc.MockCometRPC) {
				m.EXPECT().BroadcastTxSync(context.Background(), gomock.Any()).Return(&coretypes.ResultBroadcastTx{
					Code:      0,
					Data:      []byte{},
					Log:       "",
					Codespace: "",
					Hash:      []byte("%�����\u0010\n�T�\u0017\u0016�N^H[5�\u0006}�n�w�/Vi� "),
				}, nil)
			},
		},
		{
			name: "async",
			mode: BroadcastAsync,
			setupMock: func(m *mockrpc.MockCometRPC) {
				m.EXPECT().BroadcastTxAsync(context.Background(), gomock.Any()).Return(&coretypes.ResultBroadcastTx{
					Code:      0,
					Data:      []byte{},
					Log:       "",
					Codespace: "",
					Hash:      []byte("%�����\u0010\n�T�\u0017\u0016�N^H[5�\u0006}�n�w�/Vi� "),
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.mode = tt.mode
			tt.setupMock(cometMock)
			got, err := c.Broadcast(context.Background(), []byte{})
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NotNil(t, got)
			}
		})
	}
}

func Test_checkCometError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want *apiacbci.TxResponse
	}{
		{
			name: "tx already in cache",
			err:  errors.New("tx already exists in cache"),
			want: &apiacbci.TxResponse{
				Code: 19,
			},
		},
		{
			name: "mempool is full",
			err:  mempool.ErrMempoolIsFull{},
			want: &apiacbci.TxResponse{
				Code: 20,
			},
		},
		{
			name: "tx too large",
			err:  mempool.ErrTxTooLarge{},
			want: &apiacbci.TxResponse{
				Code: 21,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkCometError(tt.err, []byte{})
			require.Equal(t, got.Code, tt.want.Code)
		})
	}
}
