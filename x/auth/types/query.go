package types

import codectypes "github.com/Stride-Labs/cosmos-sdk/codec/types"

func (m *QueryAccountResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var account AccountI
	return unpacker.UnpackAny(m.Account, &account)
}

var _ codectypes.UnpackInterfacesMessage = &QueryAccountResponse{}
