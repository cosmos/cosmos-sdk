package init

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	defaultAmount                  = "100steak"
	defaultCommissionRate          = "0.1"
	defaultCommissionMaxRate       = "0.2"
	defaultCommissionMaxChangeRate = "0.01"
)

// GenTxCmd builds the gaiad gentx command.
// nolint: errcheck
func GenTxCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gentx",
		Short: "Generate a genesis tx carrying a self delegation",
		Long: fmt.Sprintf(`This command is an alias of the 'gaiad tx create-validator' command'.

It creates a genesis piece carrying a self delegation with the
following delegation and commission default parameters:

	delegation amount:           %s
	commission rate:             %s
	commission max rate:         %s
	commission max change rate:  %s
`, defaultAmount, defaultCommissionRate, defaultCommissionMaxRate, defaultCommissionMaxChangeRate),
		RunE: func(cmd *cobra.Command, args []string) error {

			config := ctx.Config
			config.SetRoot(viper.GetString(tmcli.HomeFlag))
			nodeID, valPubKey, err := InitializeNodeValidatorFiles(ctx.Config)
			if err != nil {
				return err
			}
			ip, err := server.ExternalIP()
			if err != nil {
				return err
			}

			// Run gaiad tx create-validator
			prepareFlagsForTxCreateValidator(config, nodeID, ip, valPubKey)
			createValidatorCmd := cli.GetCmdCreateValidator(cdc)

			w, err := ioutil.TempFile("", "gentx")
			if err != nil {
				return err
			}
			unsignedGenTxFilename := w.Name()
			defer os.Remove(unsignedGenTxFilename)
			os.Stdout = w
			if err = createValidatorCmd.RunE(nil, args); err != nil {
				return err
			}
			w.Close()

			prepareFlagsForTxSign()
			signCmd := authcmd.GetSignCommand(cdc, authcmd.GetAccountDecoder(cdc))
			if w, err = prepareOutputFile(config.RootDir, nodeID); err != nil {
				return err
			}
			os.Stdout = w
			return signCmd.RunE(nil, []string{unsignedGenTxFilename})
		},
	}

	cmd.Flags().String(flagClientHome, app.DefaultCLIHome, "client's home directory")
	cmd.Flags().String(client.FlagChainID, "", "genesis file chain-id")
	cmd.Flags().String(client.FlagName, "", "name of private key with which to sign the gentx")
	cmd.MarkFlagRequired(client.FlagName)
	return cmd
}

func prepareFlagsForTxCreateValidator(config *cfg.Config, nodeID, ip string, valPubKey crypto.PubKey) {
	viper.Set(tmcli.HomeFlag, viper.GetString(flagClientHome))     // --home
	viper.Set(client.FlagFrom, viper.GetString(client.FlagName))   // --from
	viper.Set(cli.FlagNodeID, nodeID)                              // --node-id
	viper.Set(cli.FlagIP, ip)                                      // --ip
	viper.Set(cli.FlagPubKey, sdk.MustBech32ifyConsPub(valPubKey)) // --pubkey
	viper.Set(cli.FlagAmount, defaultAmount)                       // --amount
	viper.Set(cli.FlagCommissionRate, defaultCommissionRate)
	viper.Set(cli.FlagCommissionMaxRate, defaultCommissionMaxRate)
	viper.Set(cli.FlagCommissionMaxChangeRate, defaultCommissionMaxChangeRate)
	viper.Set(cli.FlagGenesisFormat, true)     // --genesis-format
	viper.Set(cli.FlagMoniker, config.Moniker) // --moniker
	if config.Moniker == "" {
		viper.Set(cli.FlagMoniker, viper.GetString(client.FlagName))
	}
}

func prepareFlagsForTxSign() {
	viper.Set("offline", true)
}

func prepareOutputFile(rootDir, nodeID string) (w *os.File, err error) {
	writePath := filepath.Join(rootDir, "config", "gentx")
	if err = common.EnsureDir(writePath, 0700); err != nil {
		return
	}
	filename := filepath.Join(writePath, fmt.Sprintf("gentx-%v.json", nodeID))
	return os.Create(filename)
}
