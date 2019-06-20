package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	extypes "github.com/cosmos/cosmos-sdk/contrib/export/types"
	"github.com/cosmos/cosmos-sdk/contrib/export/v036"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
)

var migrationMap = extypes.MigrationMap{
	"v0.36": v036.Migrate,
}

const (
	flagGenesisTime = "time"
	flagChainId     = "chain-id"
)

func migrateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis [-version] [genesis.json]",
		Short: "Migrate genesis to a specified version",
		Long: strings.TrimSpace(`Migrate the source genesis into the target version and export it as standard output.

Example:
$ genesis TODO...
`),
		Args: cobra.ExactArgs(2),
		RunE: runMigrateCmd,
	}

	cmd.Flags().String(flagGenesisTime, "", "Override genesis_time with this flag")
	cmd.Flags().String(flagChainId, "", "Override chain_id with this flag")

	return cmd
}

func runMigrateCmd(cmd *cobra.Command, args []string) (err error) {
	target := args[0]
	importGenesis := args[1]

	cdc := codec.New()
	codec.RegisterCrypto(cdc)

	genDoc, err := types.GenesisDocFromFile(importGenesis)
	if err != nil {
		return err
	}

	var initialState extypes.AppMap
	cdc.MustUnmarshalJSON(genDoc.AppState, &initialState)

	if migrationMap[target] == nil {
		return fmt.Errorf("unknown migration function version: %s", target)
	}

	newGenState := migrationMap[target](initialState, cdc)
	genDoc.AppState = cdc.MustMarshalJSON(newGenState)

	genesisTime := cmd.Flag(flagGenesisTime).Value.String()
	if genesisTime != "" {
		var t time.Time

		err := t.UnmarshalText([]byte(genesisTime))
		if err != nil {
			return err
		}

		genDoc.GenesisTime = t
	}

	chainId := cmd.Flag(flagChainId).Value.String()
	if chainId != "" {
		genDoc.ChainID = chainId
	}

	out, err := cdc.MarshalJSONIndent(genDoc, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}

func main() {
	var rootCmd = &cobra.Command{Use: "migrate"}
	rootCmd.AddCommand(migrateGenesisCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
