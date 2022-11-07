package types

import autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

var AutoCLIOptions = &autocliv1.ModuleOptions{
	Tx: &autocliv1.ServiceCommandDescriptor{Service: _Msg_serviceDesc.ServiceName},
}
