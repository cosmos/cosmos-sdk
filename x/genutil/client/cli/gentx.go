package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	kbkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

// helpers for message building gen-tx command
type StakingMsgBuildingHelpers interface {
	CreateValidatorMsgHelpers(ipDefault string) (fs *flag.FlagSet, nodeIDFlag, pubkeyFlag, amountFlag, defaultsDesc string)
	PrepareFlagsForTxCreateValidator(config *cfg.Config, nodeID, chainID string, valPubKey crypto.PubKey)
	BuildCreateValidatorMsg(cliCtx context.CLIContext, txBldr auth.TxBuilder) (auth.TxBuilder, sdk.Msg, error)
}

// GenTxCmd builds the application's gentx command.
// nolint: errcheck
func GenTxCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager, smbh StakingMsgBuildingHelpers,
	genAccIterator genutil.GenesisAccountsIterator, defaultNodeHome, defaultCLIHome string) *cobra.Command {

	ipDefault, _ := server.ExternalIP()
	fsCreateValidator, flagNodeID, flagPubKey, flagAmount, defaultsDesc := smbh.CreateValidatorMsgHelpers(ipDefault)

	cmd := &cobra.Command{
		Use:   "gentx",
		Short: "Generate a genesis tx carrying a self delegation",
		Args:  cobra.NoArgs,
		Long: fmt.Sprintf(`This command is an alias of the 'tx create-validator' command'.

		It creates a genesis transaction to create a validator. 
		The following default parameters are included: 
		    %s`, defaultsDesc),

		RunE: func(cmd *cobra.Command, args []string) error {

			config := ctx.Config
			config.SetRoot(viper.GetString(client.FlagHome))
			nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(ctx.Config)
			if err != nil {
				return err
			}

			// Read --nodeID, if empty take it from priv_validator.json
			if nodeIDString := viper.GetString(flagNodeID); nodeIDString != "" {
				nodeID = nodeIDString
			}
			// Read --pubkey, if empty take it from priv_validator.json
			if valPubKeyString := viper.GetString(flagPubKey); valPubKeyString != "" {
				valPubKey, err = sdk.GetConsPubKeyBech32(valPubKeyString)
				if err != nil {
					return err
				}
			}

			genDoc, err := tmtypes.GenesisDocFromFile(config.GenesisFile())
			if err != nil {
				return err
			}

			var genesisState map[string]json.RawMessage
			if err = cdc.UnmarshalJSON(genDoc.AppState, &genesisState); err != nil {
				return err
			}

			if err = mbm.ValidateGenesis(genesisState); err != nil {
				return err
			}

			kb, err := client.NewKeyBaseFromDir(viper.GetString(flagClientHome))
			if err != nil {
				return err
			}

			name := viper.GetString(client.FlagName)
			key, err := kb.Get(name)
			if err != nil {
				return err
			}

			// Set flags for creating gentx
			viper.Set(client.FlagHome, viper.GetString(flagClientHome))
			smbh.PrepareFlagsForTxCreateValidator(config, nodeID, genDoc.ChainID, valPubKey)

			// Fetch the amount of coins staked
			amount := viper.GetString(flagAmount)
			coins, err := sdk.ParseCoins(amount)
			if err != nil {
				return err
			}

			err = genutil.ValidateAccountInGenesis(genesisState, genAccIterator, key.GetAddress(), coins, cdc)
			if err != nil {
				return err
			}

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := client.NewCLIContext().WithCodec(cdc)

			// Set the generate-only flag here after the CLI context has
			// been created. This allows the from name/key to be correctly populated.
			//
			// TODO: Consider removing the manual setting of generate-only in
			// favor of a 'gentx' flag in the create-validator command.
			viper.Set(client.FlagGenerateOnly, true)

			// create a 'create-validator' message
			txBldr, msg, err := smbh.BuildCreateValidatorMsg(cliCtx, txBldr)
			if err != nil {
				return err
			}

			info, err := txBldr.Keybase().Get(name)
			if err != nil {
				return err
			}

			if info.GetType() == kbkeys.TypeOffline || info.GetType() == kbkeys.TypeMulti {
				fmt.Println("Offline key passed in. Use `tx sign` command to sign:")
				return utils.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg})
			}

			// write the unsigned transaction to the buffer
			w := bytes.NewBuffer([]byte{})
			cliCtx = cliCtx.WithOutput(w)

			if err = utils.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg}); err != nil {
				return err
			}

			// read the transaction
			stdTx, err := readUnsignedGenTxFile(cdc, w)
			if err != nil {
				return err
			}

			// sign the transaction and write it to the output file
			signedTx, err := utils.SignStdTx(txBldr, cliCtx, name, stdTx, false, true)
			if err != nil {
				return err
			}

			// Fetch output file name
			outputDocument := viper.GetString(client.FlagOutputDocument)
			if outputDocument == "" {
				outputDocument, err = makeOutputFilepath(config.RootDir, nodeID)
				if err != nil {
					return err
				}
			}

			if err := writeSignedGenTx(cdc, outputDocument, signedTx); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Genesis transaction written to %q\n", outputDocument)
			return nil

		},
	}

	cmd.Flags().String(client.FlagHome, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultCLIHome, "client's home directory")
	cmd.Flags().String(client.FlagName, "", "name of private key with which to sign the gentx")
	cmd.Flags().String(client.FlagOutputDocument, "",
		"write the genesis transaction JSON document to the given file instead of the default location")
	cmd.Flags().AddFlagSet(fsCreateValidator)

	cmd.MarkFlagRequired(client.FlagName)
	return cmd
}

func makeOutputFilepath(rootDir, nodeID string) (string, error) {
	writePath := filepath.Join(rootDir, "config", "gentx")
	if err := common.EnsureDir(writePath, 0700); err != nil {
		return "", err
	}
	return filepath.Join(writePath, fmt.Sprintf("gentx-%v.json", nodeID)), nil
}

func readUnsignedGenTxFile(cdc *codec.Codec, r io.Reader) (auth.StdTx, error) {
	var stdTx auth.StdTx
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return stdTx, err
	}
	err = cdc.UnmarshalJSON(bytes, &stdTx)
	return stdTx, err
}

func writeSignedGenTx(cdc *codec.Codec, outputDocument string, tx auth.StdTx) error {
	outputFile, err := os.OpenFile(outputDocument, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	json, err := cdc.MarshalJSON(tx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(outputFile, "%s\n", json)
	return err
}

// DONTCOVER
