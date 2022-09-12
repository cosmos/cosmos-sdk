package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const chainUpgradeGuide = "https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md"

// ValidateGenesisCmd takes a genesis file, and makes sure that it is valid.
func ValidateGenesisCmd(mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-genesis [file]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "validates the genesis file at the default location or at the location passed as an arg",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx := client.GetClientContextFromCmd(cmd)

			cdc := clientCtx.Codec

			// Load default if passed no args, otherwise load passed file
			var genesis string
			if len(args) == 0 {
				genesis = serverCtx.Config.GenesisFile()
			} else {
				genesis = args[0]
			}

			genesisFilePath, err := cmd.Flags().GetString(flags.FlagGenesisFilePath)
			if err != nil {
				return err
			}

			if len(genesisFilePath) > 0 {
				genDoc, err := validateGenDoc(filepath.Join(genesisFilePath, "genesis.json"))
				if err != nil {
					return err
				}

				jsonObj := make(map[string]json.RawMessage)
				jsonObj["module_genesis_state"] = []byte("true")
				loadAppStateFromFolder, _ := json.Marshal(jsonObj)

				if bytes.Equal(genDoc.AppState, loadAppStateFromFolder) {
					return fmt.Errorf("genesisAppState is not equal to expectedAppState, expect: %v, actual: %v", loadAppStateFromFolder, genDoc.AppState)
				}

				for _, b := range mbm {
					bz, err := module.FileRead(filepath.Join(genesisFilePath, b.Name()), b.Name())
					if err != nil {
						return err
					}

					if err = b.ValidateGenesis(cdc, clientCtx.TxConfig, bz); err != nil {
						return fmt.Errorf("error validating genesis state in module %s: %v", b.Name(), err.Error())
					}
				}

				fmt.Printf("The genesis in %s is valid\n", genesisFilePath)
			} else {
				genDoc, err := validateGenDoc(genesis)
				if err != nil {
					return err
				}

				var genState map[string]json.RawMessage
				if err = json.Unmarshal(genDoc.AppState, &genState); err != nil {
					return fmt.Errorf("error unmarshalling genesis doc %s: %s", genesis, err.Error())
				}

				if err = mbm.ValidateGenesis(cdc, clientCtx.TxConfig, genState); err != nil {
					return fmt.Errorf("error validating genesis file %s: %s", genesis, err.Error())
				}
				fmt.Printf("File at %s is a valid genesis file\n", genesis)
			}

			return nil
		},
	}

	cmd.Flags().String(flags.FlagGenesisFilePath, "", "the file path of app genesis states")

	return cmd
}

// validateGenDoc reads a genesis file and validates that it is a correct
// Tendermint GenesisDoc. This function does not do any cosmos-related
// validation.
func validateGenDoc(importGenesisFile string) (*tmtypes.GenesisDoc, error) {
	genDoc, err := tmtypes.GenesisDocFromFile(importGenesisFile)
	if err != nil {
		return nil, fmt.Errorf("%s. Make sure that"+
			" you have correctly migrated all Tendermint consensus params, please see the"+
			" chain migration guide at %s for more info",
			err.Error(), chainUpgradeGuide,
		)
	}

	return genDoc, nil
}
