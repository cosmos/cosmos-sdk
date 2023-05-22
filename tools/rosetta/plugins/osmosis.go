package main

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
)

func InitZone() {

	fmt.Println("Init Osmosis Zone")

	config := types.GetConfig()

	prefix := "osmo"
	config.SetBech32PrefixForAccount(prefix, prefix+"pub")
	config.SetBech32PrefixForValidator(prefix+"valoper", prefix+"valoperpub")
	config.SetBech32PrefixForConsensusNode(prefix+"valcons", prefix+"valconspub")
}
