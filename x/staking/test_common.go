package staking

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// nolint: deadcode unused
var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = secp256k1.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())
	addr3 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	priv4 = secp256k1.GenPrivKey()
	addr4 = sdk.AccAddress(priv4.PubKey().Address())
	coins = sdk.Coins{sdk.NewCoin("foocoin", sdk.NewInt(10))}
	fee   = auth.NewStdFee(
		100000,
		sdk.Coins{sdk.NewCoin("foocoin", sdk.NewInt(0))},
	)

	commissionRates = NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
)

func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey crypto.PubKey, amt sdk.Int) MsgCreateValidator {
	return types.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(sdk.DefaultBondDenom, amt), Description{}, commissionRates, sdk.OneInt(),
	)
}

func NewTestMsgCreateValidatorWithCommission(address sdk.ValAddress, pubKey crypto.PubKey,
	amt sdk.Int, commissionRate sdk.Dec) MsgCreateValidator {

	commission := NewCommissionRates(commissionRate, sdk.OneDec(), sdk.ZeroDec())

	return types.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(sdk.DefaultBondDenom, amt), Description{}, commission, sdk.OneInt(),
	)
}

func NewTestMsgCreateValidatorWithMinSelfDelegation(address sdk.ValAddress, pubKey crypto.PubKey,
	amt sdk.Int, minSelfDelegation sdk.Int) MsgCreateValidator {

	return types.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(sdk.DefaultBondDenom, amt), Description{}, commissionRates, minSelfDelegation,
	)
}

func NewTestMsgDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Int) MsgDelegate {
	amount := sdk.NewCoin(sdk.DefaultBondDenom, amt)
	return NewMsgDelegate(delAddr, valAddr, amount)
}
