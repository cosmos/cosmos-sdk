package staking_test

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// nolint:deadcode,unused,varcheck
var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = secp256k1.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())

	commissionRates = staking.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
)

func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey crypto.PubKey, amt sdk.Int) staking.MsgCreateValidator {
	return types.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(sdk.DefaultBondDenom, amt), staking.Description{}, commissionRates, sdk.OneInt(),
	)
}

func NewTestMsgDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Int) staking.MsgDelegate {
	amount := sdk.NewCoin(sdk.DefaultBondDenom, amt)
	return staking.NewMsgDelegate(delAddr, valAddr, amount)
}
