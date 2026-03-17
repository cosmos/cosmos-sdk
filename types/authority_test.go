package types_test

import (
	"errors"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func TestValidateAuthority(t *testing.T) {
	key := storetypes.NewKVStoreKey(t.Name())
	ctx := testutil.DefaultContext(key, storetypes.NewTransientStoreKey("transient_"+t.Name()))

	keeperAuth := "keeper-authority"
	consAuth := "consensus-authority"

	// No consensus params — falls back to keeper authority.
	require.NoError(t, types.ValidateAuthority(ctx, keeperAuth, keeperAuth))

	// No consensus params — wrong msg authority fails with ErrUnauthorized.
	err := types.ValidateAuthority(ctx, keeperAuth, "wrong")
	require.Error(t, err)
	require.True(t, errors.Is(err, sdkerrors.ErrUnauthorized))
	require.Contains(t, err.Error(), "invalid authority")
	require.Contains(t, err.Error(), keeperAuth)

	// Consensus params authority set — takes precedence over keeper.
	ctx = ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Authority: &cmtproto.AuthorityParams{Authority: consAuth},
	})
	require.NoError(t, types.ValidateAuthority(ctx, keeperAuth, consAuth))

	// Consensus params authority set — keeper authority is rejected.
	err = types.ValidateAuthority(ctx, keeperAuth, keeperAuth)
	require.Error(t, err)
	require.True(t, errors.Is(err, sdkerrors.ErrUnauthorized))
	require.Contains(t, err.Error(), consAuth)

	// Consensus params with empty authority — falls back to keeper.
	ctx = ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Authority: &cmtproto.AuthorityParams{Authority: ""},
	})
	require.NoError(t, types.ValidateAuthority(ctx, keeperAuth, keeperAuth))

	// Consensus params with nil AuthorityParams — falls back to keeper.
	ctx = ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Authority: nil,
	})
	require.NoError(t, types.ValidateAuthority(ctx, keeperAuth, keeperAuth))
}
