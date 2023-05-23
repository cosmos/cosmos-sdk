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

	fmt.Println("Init Osmosis Zone")

	config := sdk.GetConfig()

	prefix := "osmo"
	config.SetBech32PrefixForAccount(prefix, prefix+"pub")
	config.SetBech32PrefixForValidator(prefix+"valoper", prefix+"valoperpub")
	config.SetBech32PrefixForConsensusNode(prefix+"valcons", prefix+"valconspub")
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	fmt.Println("Registering Osmosis interfaces")

	ibcclienttypes.RegisterInterfaces(registry)
	ibcLightClient.RegisterInterfaces(registry)
	sdk.RegisterInterfaces(registry)
	txtypes.RegisterInterfaces(registry)
	cryptocodec.RegisterInterfaces(registry)
}
