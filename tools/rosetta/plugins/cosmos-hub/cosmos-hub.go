package main

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	ibcLightClient "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func InitZone() {

	fmt.Println("Init Cosmos-hub Zone")

	config := sdk.GetConfig()

	prefix := "cosmos"
	config.SetBech32PrefixForAccount(prefix, prefix+"pub")
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	fmt.Println("Registering Osmosis interfaces")

	ibcclienttypes.RegisterInterfaces(registry)
	ibcLightClient.RegisterInterfaces(registry)
	sdk.RegisterInterfaces(registry)
	txtypes.RegisterInterfaces(registry)
	cryptocodec.RegisterInterfaces(registry)
	
	
}
