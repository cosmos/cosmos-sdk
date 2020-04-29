package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests/cli/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashcli "github.com/cosmos/cosmos-sdk/x/slashing/client/cli_test"
)

func TestSlashingGetParams(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	defer proc.Stop(false)

	params := slashcli.QuerySlashingParams(f)
	require.Equal(t, int64(100), params.SignedBlocksWindow)
	require.Equal(t, sdk.NewDecWithPrec(5, 1), params.MinSignedPerWindow)

	sinfo := slashcli.QuerySigningInfo(f, f.SDTendermint("show-validator"))
	require.Equal(t, int64(0), sinfo.StartHeight)
	require.False(t, sinfo.Tombstoned)

	// Cleanup testing directories
	f.Cleanup()
}
