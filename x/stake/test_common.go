package stake

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
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
		sdk.Coins{sdk.NewCoin("foocoin", sdk.NewInt(0))}...,
	)

	commissionMsg = NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
)

func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey crypto.PubKey, amt int64) MsgCreateValidator {
	return types.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(types.DefaultBondDenom, sdk.NewInt(amt)), Description{}, commissionMsg,
	)
}

func NewTestMsgCreateValidatorWithCommission(address sdk.ValAddress, pubKey crypto.PubKey,
	amt int64, commissionRate sdk.Dec) MsgCreateValidator {

	commission := NewCommissionMsg(commissionRate, sdk.OneDec(), sdk.ZeroDec())

	return types.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(types.DefaultBondDenom, sdk.NewInt(amt)), Description{}, commission,
	)
}

func NewTestMsgDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt int64) MsgDelegate {
	return MsgDelegate{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		Delegation:    sdk.NewCoin(types.DefaultBondDenom, sdk.NewInt(amt)),
	}
}

func NewTestMsgCreateValidatorOnBehalfOf(delAddr sdk.AccAddress, valAddr sdk.ValAddress, valPubKey crypto.PubKey, amt int64) MsgCreateValidator {
	return MsgCreateValidator{
		Description:   Description{},
		Commission:    commissionMsg,
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		PubKey:        valPubKey,
		Delegation:    sdk.NewCoin(types.DefaultBondDenom, sdk.NewInt(amt)),
	}
}
