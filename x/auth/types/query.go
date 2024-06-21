package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprotoany "github.com/cosmos/gogoproto/types/any"
)

func (m *QueryAccountResponse) UnpackInterfaces(unpacker gogoprotoany.AnyUnpacker) error {
	var account sdk.AccountI
	return unpacker.UnpackAny(m.Account, &account)
}

var _ gogoprotoany.UnpackInterfacesMessage = &QueryAccountResponse{}
