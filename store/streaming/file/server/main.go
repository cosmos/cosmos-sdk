package main

import (
	"github.com/rs/zerolog"

	"github.com/cosmos/cosmos-sdk/store/streaming/file/server/cmd"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	cmd.Execute()
}
