package types

import paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

// Parameter keys
var (
	ParamStoreKeyCommunityTax        = []byte("communitytax")
	ParamStoreKeyWithdrawAddrEnabled = []byte("withdrawaddrenabled")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyCommunityTax, &p.CommunityTax, validateCommunityTax),
		paramtypes.NewParamSetPair(ParamStoreKeyWithdrawAddrEnabled, &p.WithdrawAddrEnabled, validateWithdrawAddrEnabled),
	}
}
