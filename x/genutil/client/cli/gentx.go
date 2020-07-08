package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

// GenTxCmd builds the application's gentx command.
// nolint: errcheck
func GenTxCmd(mbm module.BasicManager, genBalIterator types.GenesisBalancesIterator, defaultNodeHome string) *cobra.Command {
	ipDefault, _ := server.ExternalIP()
	fsCreateValidator, defaultsDesc := cli.CreateValidatorMsgFlagSet(ipDefault)

	cmd := &cobra.Command{
		Use:   "gentx",
		Short: "Generate a genesis tx carrying a self delegation",
		Args:  cobra.NoArgs,
		Long: fmt.Sprintf(`This command is an alias of the 'tx create-validator' command'.

		It creates a genesis transaction to create a validator. 
		The following default parameters are included: 
		    %s`, defaultsDesc),

		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.JSONMarshaler

			home, _ := cmd.Flags().GetString(flags.FlagHome)

			config := serverCtx.Config
			config.SetRoot(home)

			nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(serverCtx.Config)
			if err != nil {
				return errors.Wrap(err, "failed to initialize node validator files")
			}

			// Read --nodeID, if empty take it from priv_validator.json
			nodeIDString, _ := cmd.Flags().GetString(cli.FlagNodeID)

			if nodeIDString != "" {
				nodeID = nodeIDString
			}
			// Read --pubkey, if empty take it from priv_validator.json
			valPubKeyString, _ := cmd.Flags().GetString(cli.FlagPubKey)

			if valPubKeyString != "" {
				valPubKey, err = sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, valPubKeyString)
				if err != nil {
					return errors.Wrap(err, "failed to get consensus node public key")
				}
			}

			genDoc, err := tmtypes.GenesisDocFromFile(config.GenesisFile())
			if err != nil {
				return errors.Wrapf(err, "failed to read genesis doc file %s", config.GenesisFile())
			}

			var genesisState map[string]json.RawMessage
			if err = cdc.UnmarshalJSON(genDoc.AppState, &genesisState); err != nil {
				return errors.Wrap(err, "failed to unmarshal genesis state")
			}

			if err = mbm.ValidateGenesis(cdc, genesisState); err != nil {
				return errors.Wrap(err, "failed to validate genesis state")
			}

			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf)
			if err != nil {
				return errors.Wrap(err, "failed to initialize keybase")
			}

			name, _ := cmd.Flags().GetString(flags.FlagName)

			key, err := kb.Key(name)
			if err != nil {
				return errors.Wrap(err, "failed to read from keybase")
			}

			// Set flags for creating gentx
			createValCfg, err := cli.PrepareConfigForTxCreateValidator(config, cmd.Flags(), nodeID, genDoc.ChainID, valPubKey)
			if err != nil {
				return errors.Wrap(err, "error creating configuration to create validator msg")
			}

			// Fetch the amount of coins staked
			amount, _ := cmd.Flags().GetString(cli.FlagAmount)

			coins, err := sdk.ParseCoins(amount)
			if err != nil {
				return errors.Wrap(err, "failed to parse coins")
			}

			err = genutil.ValidateAccountInGenesis(genesisState, genBalIterator, key.GetAddress(), coins, cdc)
			if err != nil {
				return errors.Wrap(err, "failed to validate account in genesis")
			}

			txBldr, err := authtypes.NewTxBuilderFromFlags(inBuf, cmd.Flags(), clientCtx.HomeDir)
			if err != nil {
				return errors.Wrap(err, "error creating tx builder")
			}

			txBldr = txBldr.WithTxEncoder(authclient.GetTxEncoder(clientCtx.Codec))

			from, _ := cmd.Flags().GetString(flags.FlagFrom)
			fromAddress, _, err := client.GetFromFields(txBldr.Keybase(), from, false)
			if err != nil {
				return errors.Wrap(err, "error getting from address")
			}

			clientCtx = clientCtx.WithInput(inBuf).WithFromAddress(fromAddress)

			// create a 'create-validator' message
			txBldr, msg, err := cli.BuildCreateValidatorMsg(clientCtx, createValCfg, txBldr, true)
			if err != nil {
				return errors.Wrap(err, "failed to build create-validator message")
			}

			if key.GetType() == keyring.TypeOffline || key.GetType() == keyring.TypeMulti {
				cmd.PrintErrln("Offline key passed in. Use `tx sign` command to sign.")
				return authclient.PrintUnsignedStdTx(txBldr, clientCtx, []sdk.Msg{msg})
			}

			// write the unsigned transaction to the buffer
			w := bytes.NewBuffer([]byte{})
			clientCtx = clientCtx.WithOutput(w)

			if err = authclient.PrintUnsignedStdTx(txBldr, clientCtx, []sdk.Msg{msg}); err != nil {
				return errors.Wrap(err, "failed to print unsigned std tx")
			}

			// read the transaction
			stdTx, err := readUnsignedGenTxFile(cdc, w)
			if err != nil {
				return errors.Wrap(err, "failed to read unsigned gen tx file")
			}

			// sign the transaction and write it to the output file
			signedTx, err := authclient.SignStdTx(txBldr, clientCtx, name, stdTx, false, true)
			if err != nil {
				return errors.Wrap(err, "failed to sign std tx")
			}

			// Fetch output file name
			outputDocument, _ := cmd.Flags().GetString(flags.FlagOutputDocument)

			if outputDocument == "" {
				outputDocument, err = makeOutputFilepath(config.RootDir, nodeID)
				if err != nil {
					return errors.Wrap(err, "failed to create output file path")
				}
			}

			if err := writeSignedGenTx(cdc, outputDocument, signedTx); err != nil {
				return errors.Wrap(err, "failed to write signed gen tx")
			}

			cmd.PrintErrf("Genesis transaction written to %q\n", outputDocument)
			return nil

		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagName, "", "name of private key with which to sign the gentx")
	cmd.Flags().String(flags.FlagOutputDocument, "", "write the genesis transaction JSON document to the given file instead of the default location")
	cmd.Flags().AddFlagSet(fsCreateValidator)
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.MarkFlagRequired(flags.FlagName)

	flags.PostCommands(cmd)

	return cmd
}

func makeOutputFilepath(rootDir, nodeID string) (string, error) {
	writePath := filepath.Join(rootDir, "config", "gentx")
	if err := tmos.EnsureDir(writePath, 0700); err != nil {
		return "", err
	}

	return filepath.Join(writePath, fmt.Sprintf("gentx-%v.json", nodeID)), nil
}

func readUnsignedGenTxFile(cdc codec.JSONMarshaler, r io.Reader) (authtypes.StdTx, error) {
	var stdTx authtypes.StdTx

	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return stdTx, err
	}

	err = cdc.UnmarshalJSON(bytes, &stdTx)
	return stdTx, err
}

func writeSignedGenTx(cdc codec.JSONMarshaler, outputDocument string, tx authtypes.StdTx) error {
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
