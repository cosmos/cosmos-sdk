package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestConsensusParamsRejectPubKeyTypeInUse checks that depinject wires x/staking into
// x/consensus: removing the key type of the genesis validator is rejected.
func TestConsensusParamsRejectPubKeyTypeInUse(t *testing.T) {
	var consensusKeeper consensuskeeper.Keeper
	app, err := simtestutil.Setup(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.BankModule(),
				configurator.StakingModule(),
				configurator.TxModule(),
				configurator.ConsensusModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&consensusKeeper,
	)
	require.NoError(t, err)

	ctx := app.NewContext(false)
	params, err := consensusKeeper.ParamsStore.Get(ctx)
	require.NoError(t, err)

	updateParams := func(pubKeyTypes ...string) error {
		_, err := consensusKeeper.UpdateParams(ctx, &consensustypes.MsgUpdateParams{
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			Block:     params.Block,
			Evidence:  params.Evidence,
			Validator: &cmtproto.ValidatorParams{PubKeyTypes: pubKeyTypes},
		})
		return err
	}

	// the genesis validator uses an ed25519 consensus key
	err = updateParams(cmttypes.ABCIPubKeyTypeSecp256k1)
	require.ErrorIs(t, err, stakingtypes.ErrValidatorPubKeyTypeNotSupported)

	got, err := consensusKeeper.ParamsStore.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, params.Validator, got.Validator)

	require.NoError(t, updateParams(cmttypes.ABCIPubKeyTypeEd25519, cmttypes.ABCIPubKeyTypeSecp256k1))
}
