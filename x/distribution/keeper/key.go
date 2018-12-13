package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// keys/key-prefixes
var (
	FeePoolKey  = []byte{0x00} // key for global distribution state
	ProposerKey = []byte{0x04} // key for storing the proposer operator address

	DelegatorWithdrawAddrKey = []byte{0x05} // key for storing delegator withdraw address

	// params store
	ParamStoreKeyCommunityTax        = []byte("communitytax")
	ParamStoreKeyBaseProposerReward  = []byte("baseproposerreward")
	ParamStoreKeyBonusProposerReward = []byte("bonusproposerreward")
)

const (
	// default paramspace for params keeper
	DefaultParamspace = "distr"
)

// gets the prefix for a delegator's withdraw info
func GetDelegatorWithdrawAddrKey(delAddr sdk.AccAddress) []byte {
	return append(DelegatorWithdrawAddrKey, delAddr.Bytes()...)
}

// gets an address from a delegator's withdraw info key
func GetDelegatorWithdrawInfoAddress(key []byte) (delAddr sdk.AccAddress) {
	addr := key[1:]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	return sdk.AccAddress(addr)
}
