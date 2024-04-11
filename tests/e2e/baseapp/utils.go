package baseapp

import (
	"encoding/json"
	"testing"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	_ "cosmossdk.io/x/auth"
	_ "cosmossdk.io/x/auth/tx/config"
	authtypes "cosmossdk.io/x/auth/types"
	_ "cosmossdk.io/x/bank"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/staking"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
)

// GenesisStateWithSingleValidator initializes GenesisState with a single validator and genesis accounts
// that also act as delegators.
func GenesisStateWithSingleValidator(t *testing.T, codec codec.Codec, builder *runtime.AppBuilder) map[string]json.RawMessage {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balances := []banktypes.Balance{
		{
			Address: acc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000000000000))),
		},
	}

	genesisState := builder.DefaultGenesis()
	// sus
	genesisState, err = simtestutil.GenesisStateWithValSet(codec, genesisState, valSet, []authtypes.GenesisAccount{acc}, balances...)
	require.NoError(t, err)

	return genesisState
}
