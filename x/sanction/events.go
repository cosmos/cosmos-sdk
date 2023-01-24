package sanction

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewEventAddressSanctioned(addr sdk.AccAddress) *EventAddressSanctioned {
	return &EventAddressSanctioned{
		Address: addr.String(),
	}
}

func NewEventAddressUnsanctioned(addr sdk.AccAddress) *EventAddressUnsanctioned {
	return &EventAddressUnsanctioned{
		Address: addr.String(),
	}
}

func NewEventTempAddressSanctioned(addr sdk.AccAddress) *EventTempAddressSanctioned {
	return &EventTempAddressSanctioned{
		Address: addr.String(),
	}
}

func NewEventTempAddressUnsanctioned(addr sdk.AccAddress) *EventTempAddressUnsanctioned {
	return &EventTempAddressUnsanctioned{
		Address: addr.String(),
	}
}
