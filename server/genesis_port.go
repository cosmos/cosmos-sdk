package server

// DONTCOVER

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	gapp "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisPortCmd ports old genesis file and update its app state for a software upgrade
func GenesisPortCmd(ctx *Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis-port [old-genesis.json] [chain-id] [start-time]",
		Short: "Port old genesis file and update its app state for a software upgrade",
		Long: strings.TrimSpace(`Port old genesis file and update its app state for a software upgrade.

$ gaiad genesis-port cosmoshub-1 2019-02-11T12:00:00Z > new_genesis.json
`),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldGenFilename := args[0]
			newChainID := args[1]
			genesisTimeStr := args[2]

			bz := []byte(genesisTimeStr)
			genesisTime, err := sdk.ParseTimeBytes(bz)
			if err != nil {
				return err
			}

			if ext := filepath.Ext(oldGenFilename); ext != ".json" {
				return fmt.Errorf("%s is not a JSON file", oldGenFilename)
			}

			if _, err = os.Stat(oldGenFilename); err != nil {
				return err
			}

			genesis, err := gapp.GetUpdatedGenesis(cdc, oldGenFilename, newChainID, genesisTime)
			if err != nil {
				return err
			}

			genesisJSON, err := cdc.MarshalJSONIndent(genesis, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(genesisJSON))
			return nil
		},
	}
	return cmd
}
