package broadcast

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type testBFT struct{}

func (t testBFT) Broadcast(_ context.Context, _ []byte) ([]byte, error) {
	return nil, nil
}

func (t testBFT) Consensus() string {
	return "testBFT"
}

func Test_newBroadcaster(t *testing.T) {
	f := NewFactory()
	tests := []struct {
		name      string
		consensus string
		opts      []Option
		want      Broadcaster
		wantErr   bool
	}{
		{
			name:      "comet",
			consensus: "comet",
			opts: []Option{
				withMode(BroadcastSync),
			},
			want: &CometBFTBroadcaster{},
		},
		{
			name:      "unsupported_consensus",
			consensus: "unsupported",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.Create(context.Background(), tt.consensus, "localhost:26657", tt.opts...)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				require.IsType(t, tt.want, got)
			}
		})
	}
}

func TestFactory_Register(t *testing.T) {
	tests := []struct {
		name      string
		consensus string
		creator   NewBroadcasterFn
	}{
		{
			name:      "register new broadcaster",
			consensus: "testBFT",
			creator: func(url string, opts ...Option) (Broadcaster, error) {
				return testBFT{}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFactory()
			f.Register(tt.consensus, tt.creator)

			b, err := f.Create(context.Background(), tt.consensus, "localhost:26657")
			require.NoError(t, err)
			require.NotNil(t, b)

			require.Equal(t, tt.consensus, b.Consensus())
		})
	}
}
