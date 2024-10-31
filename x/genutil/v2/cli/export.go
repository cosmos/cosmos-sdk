package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const (
	flagHeight           = "height"
	flagJailAllowedAddrs = "jail-allowed-addrs"
)

// ExportCmd dumps app state to JSON.
func ExportCmd(app ExportableApp) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config := client.GetConfigFromCmd(cmd)

			if _, err := os.Stat(config.GenesisFile()); os.IsNotExist(err) {
				return err
			}

			if app == nil {
				if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: App exporter not defined. Returning genesis file."); err != nil {
					return err
				}

				// Open file in read-only mode so we can copy it to stdout.
				// It is possible that the genesis file is large,
				// so we don't need to read it all into memory
				// before we stream it out.
				f, err := os.OpenFile(config.GenesisFile(), os.O_RDONLY, 0)
				if err != nil {
					return err
				}
				defer f.Close()

				if _, err := io.Copy(cmd.OutOrStdout(), f); err != nil {
					return err
				}

				return nil
			}

			height, _ := cmd.Flags().GetInt64(flagHeight)
			jailAllowedAddrs, _ := cmd.Flags().GetStringSlice(flagJailAllowedAddrs)
			outputDocument, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
			if height != -1 {
				if err := app.LoadHeight(uint64(height)); err != nil {
					return err
				}
			}
			exported, err := app.ExportAppStateAndValidators(jailAllowedAddrs)
			if err != nil {
				return fmt.Errorf("error exporting state: %w", err)
			}

			appGenesis, err := genutiltypes.AppGenesisFromFile(config.GenesisFile())
			if err != nil {
				return err
			}

			// set current binary version
			appGenesis.AppName = version.AppName
			appGenesis.AppVersion = version.Version

			appGenesis.AppState = exported.AppState
			appGenesis.InitialHeight = exported.Height
			appGenesis.Consensus.Validators = exported.Validators

			out, err := json.Marshal(appGenesis)
			if err != nil {
				return err
			}

			if outputDocument == "" {
				// Copy the entire genesis file to stdout.
				_, err := io.Copy(cmd.OutOrStdout(), bytes.NewReader(out))
				return err
			}

			if err = appGenesis.SaveAs(outputDocument); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().Int64(flagHeight, -1, "Export state from a particular height (-1 means latest height)")
	cmd.Flags().
		StringSlice(flagJailAllowedAddrs, []string{}, "Comma-separated list of operator addresses of jailed validators to unjail")
	cmd.Flags().
		String(flags.FlagOutputDocument, "", "Exported state is written to the given file instead of STDOUT")

	return cmd
}
