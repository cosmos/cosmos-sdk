package keeper

// DONTCOVER

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO remove dependencies on staking (should only refer to validator set type from sdk)

var (
	InitTokens = sdk.TokensFromConsensusPower(200)
)

// Have to change these parameters for tests
// lest the tests take forever
func TestParams() types.Params {
	params := types.DefaultParams()
	params.SignedBlocksWindow = 1000
	params.DowntimeJailDuration = 60 * 60

	return params
}

// TODO: remove this
func NewTestMsgCreateValidator1(address sdk.ValAddress, pubKey crypto.PubKey, amt sdk.Int, t *testing.T) *stakingtypes.MsgCreateValidator {
	commission := stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())

	msg, err := stakingtypes.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(sdk.DefaultBondDenom, amt),
		stakingtypes.Description{}, commission, sdk.OneInt(),
	)
	require.NoError(t, err)
	return msg
}

func NewTestMsgDelegate1(delAddr sdk.AccAddress, valAddr sdk.ValAddress, delAmount sdk.Int) *stakingtypes.MsgDelegate {
	amount := sdk.NewCoin(sdk.DefaultBondDenom, delAmount)
	return stakingtypes.NewMsgDelegate(delAddr, valAddr, amount)
}
