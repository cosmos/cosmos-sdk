package broadcast

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newBroadcaster(t *testing.T) {
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
			want: &CometBftBroadcaster{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := broadcasterFactory{}.create(context.Background(), tt.consensus, "localhost:26657", tt.opts...)
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
