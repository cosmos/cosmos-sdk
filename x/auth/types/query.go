package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *QueryAccountResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var account sdk.AccountI
	return unpacker.UnpackAny(m.Account, &account)
}

var _ codectypes.UnpackInterfacesMessage = &QueryAccountResponse{}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m *QueryPubKeyHistoryResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for i := range m.Entries {
		if err := m.Entries[i].UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

var _ codectypes.UnpackInterfacesMessage = &QueryPubKeyHistoryResponse{}
