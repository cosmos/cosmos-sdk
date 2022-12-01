package main

import (
	"os"

	rosettaCmd "cosmossdk.io/tools/rosetta/cmd"
	"cosmossdk.io/tools/rosetta/lib/logger"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func main() {
	var (
		logger            = logger.NewLogger()
		interfaceRegistry = codectypes.NewInterfaceRegistry()
		cdc               = codec.NewProtoCodec(interfaceRegistry)
	)

	if err := rosettaCmd.RosettaCommand(interfaceRegistry, cdc).Execute(); err != nil {
		logger.Err(err).Msg("failed to run rosetta")
		os.Exit(1)
	}
}
