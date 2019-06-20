package main

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	extypes "github.com/cosmos/cosmos-sdk/contrib/export/types"
	"github.com/cosmos/cosmos-sdk/contrib/export/v036"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
	"log"
	"time"
)

var (
	migrationMap = extypes.MigrationMap{
		"v0.36": v036.Migrate,
	}
	//source        string
	target        string
	importGenesis string
)

const (
	// FlagSource will support multiple version upgrades
	FlagSource      = "source"
	FlagGenesisTime = "time"
	FlagChainId     = "chain-id"
)

func migrateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis [v0.34] [~/my/genesis.json]",
		Short: "Migrate genesis to a specified version",
		Long:  `Migrate the source genesis into the target version and export it as standard output.`,
		Args:  cobra.ExactArgs(2),
		RunE:  runMigrateCmd,
	}

	// TODO: add support for multiple version upgrades by looping sequentially to migration functions
	cmd.Flags().String(FlagSource, "", "[Optional] The SDK version that exported the genesis")
	cmd.Flags().String(FlagGenesisTime, "", "[Optional] Override genesis_time with this flag")
	cmd.Flags().String(FlagChainId, "", "[Optional] Override chain_id with this flag")

	return cmd
}

func runMigrateCmd(cmd *cobra.Command, args []string) (err error) {
	target = args[0]
	importGenesis = args[1]

	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	// Dump here an example genesis for CCI to test import from old types, and export to new ones
	genDoc, err := types.GenesisDocFromFile(importGenesis)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	var initialState extypes.AppMap
	cdc.MustUnmarshalJSON(genDoc.AppState, &initialState)

	if migrationMap[target] == nil {
		errMessage := fmt.Sprintf("Missing migration function for version %s", target)
		log.Println(errMessage)
		return errors.New(errMessage)
	}
	newGenState := migrationMap[target](initialState, cdc)
	genDoc.AppState = cdc.MustMarshalJSON(newGenState)

	genesisTime := cmd.Flag(FlagGenesisTime).Value.String()
	chainId := cmd.Flag(FlagChainId).Value.String()
	if genesisTime != "" {
		var t time.Time
		err := t.UnmarshalText([]byte(genesisTime))
		if err != nil {
			log.Println(err.Error())
			return err
		}
		genDoc.GenesisTime = t
	}
	if chainId != "" {
		genDoc.ChainID = chainId
	}

	out, err := cdc.MarshalJSONIndent(genDoc, "", "  ")
	if err != nil {
		log.Println(err.Error())
		return err
	}
	fmt.Println(string(out))
	return nil
}

func main() {
	var rootCmd = &cobra.Command{Use: "migrate"}
	rootCmd.AddCommand(migrateGenesisCmd())
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
