package cli

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/stretchr/testify/require"

	"github.com/spf13/viper"

	"github.com/tendermint/go-amino"
)

func TestGetBroadcastCommand_CannotBePerformedWhenOfflineFlag(t *testing.T) {
	codec := amino.NewCodec()
	cmd := GetBroadcastCommand(codec)

	viper.Set(flags.FlagOffline, true)

	err := cmd.RunE(nil, []string{})
	require.EqualError(t, err, "cannot broadcast tx with offline flag")
}
