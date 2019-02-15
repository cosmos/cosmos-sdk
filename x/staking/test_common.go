package staking

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	priv1 = ed25519.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = ed25519.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())
	addr3 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	priv4 = ed25519.GenPrivKey()
	addr4 = sdk.AccAddress(priv4.PubKey().Address())
	coins = sdk.Coins{sdk.NewCoin("foocoin", sdk.NewInt(10))}
	fee   = auth.NewStdFee(
		100000,
		sdk.Coins{sdk.NewCoin("foocoin", sdk.NewInt(0))},
	)

	commissionMsg = NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
)

func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey crypto.PubKey, amt sdk.Int) MsgCreateValidator {
	return types.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(sdk.DefaultBondDenom, amt), Description{}, commissionMsg, sdk.OneInt(),
	)
}

func NewTestMsgCreateValidatorWithCommission(address sdk.ValAddress, pubKey crypto.PubKey,
	amt sdk.Int, commissionRate sdk.Dec) MsgCreateValidator {

	commission := NewCommissionMsg(commissionRate, sdk.OneDec(), sdk.ZeroDec())

	return types.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(sdk.DefaultBondDenom, amt), Description{}, commission, sdk.OneInt(),
	)
}

func NewTestMsgCreateValidatorWithMinSelfDelegation(address sdk.ValAddress, pubKey crypto.PubKey,
	amt sdk.Int, minSelfDelegation sdk.Int) MsgCreateValidator {

	return types.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(sdk.DefaultBondDenom, amt), Description{}, commissionMsg, minSelfDelegation,
	)
}

func NewTestMsgDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Int) MsgDelegate {
	amount := sdk.NewCoin(sdk.DefaultBondDenom, amt)
	return NewMsgDelegate(delAddr, valAddr, amount)
}

func NewTestMsgCreateValidatorOnBehalfOf(delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	valPubKey crypto.PubKey, amt sdk.Int) MsgCreateValidator {

	amount := sdk.NewCoin(sdk.DefaultBondDenom, amt)
	return NewMsgCreateValidatorOnBehalfOf(delAddr, valAddr, valPubKey, amount, Description{}, commissionMsg, sdk.OneInt())
}
