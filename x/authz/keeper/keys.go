package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
)

// grantStoreKey - return authorization store key
func grantStoreKey(grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) []byte {
	return append(append(append(
		types.GrantKey,
		address.MustLengthPrefix(granter)...),
		address.MustLengthPrefix(grantee)...),
		[]byte(msgType)...,
	)
}

// addressesFromGrantStoreKey - split granter & grantee address from the authorization key
func addressesFromGrantStoreKey(key []byte) (granterAddr, granteeAddr sdk.AccAddress) {
	// key if of format:
	// 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>
	granterAddrLen := key[1] // remove prefix key
	granterAddr = sdk.AccAddress(key[2 : 2+granterAddrLen])
	granteeAddrLen := int(key[2+granterAddrLen])
	granteeAddr = sdk.AccAddress(key[3+granterAddrLen : 3+granterAddrLen+byte(granteeAddrLen)])

	return granterAddr, granteeAddr
}
