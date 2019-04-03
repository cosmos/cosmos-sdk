package main

import (
	"strings"
	app "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/scripts/genesis/export"
)

func main() {
	cdc := app.MakeCodec()

	args := flags.Args()
	if len(args) != 3 {
		panic(fmt.Errorf("please provide path, chain-id and genesis time"))
	}
	
	pathToGenesis := args[0]
	chainID := args[1]
	genesisTime := args[2]

	err := export.ValidateInputs(pathToGenesis, chainID, genesisTime)
	if err != nil {
		panic(err)
	}

	genesis, err := export.NewGenesisFile(cdc, pathToGenesis)
	if err != nil {
		panic(err)
	}

	genesis.ChainID = strings.Trim(chainID, " ")
	genesis.GenesisTime = startTime

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
		return panic(err)
	}

	genesisJSON, err := cdc.MarshalJSONIndent(genesis, "", "  ")
	if err != nil {
		return panic(err)
	}
	fmt.Println(string(genesisJSON))
}
