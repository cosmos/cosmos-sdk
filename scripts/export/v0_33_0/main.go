package main

import (
	"fmt"
	"os"

	app "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/scripts/export"
)

const chainID = "cosmoshub-2"

// Command: go run main.go [path_to_old_genesis.json] [genesis-start-time] > [path_to_new_genesis.json]
func main() {
	cdc := app.MakeCodec()

	args := os.Args
	if len(args) != 2 {
		panic(fmt.Errorf("please provide path and genesis start time"))
	}

	pathToGenesis := args[0]
	genesisTime := args[1]

	err := export.ValidateBasic(pathToGenesis, genesisTime)
	if err != nil {
		panic(err)
	}

	genesis, err := export.NewGenesisFile(cdc, pathToGenesis)
	if err != nil {
		panic(err)
	}

	genesis.ChainID = chainID
	genesis.GenesisTime = genesisTime

	// proposal #1 updates
	genesis.AppState.MintData.Params.BlocksPerYear = 4855015

	// proposal #2 updates
	genesis.ConsensusParams.Block.MaxGas = 200000
	genesis.ConsensusParams.Block.MaxBytes = 2000000

	// enable transfers
	genesis.AppState.BankData.SendEnabled = true
	genesis.AppState.DistrData.WithdrawAddrEnabled = true

	err = app.GaiaValidateGenesisState(genesis.AppState)
	if err != nil {
		panic(err)
	}

	genesisJSON, err := cdc.MarshalJSONIndent(genesis, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(genesisJSON))
}
