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
	"github.com/spf13/viper"

	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/utils"

	"github.com/cosmos/cosmos-sdk/codec"
	kbkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

var (
	defaultTokens                  = sdk.TokensFromTendermintPower(100)
	defaultAmount                  = defaultTokens.String() + sdk.DefaultBondDenom
	defaultCommissionRate          = "0.1"
	defaultCommissionMaxRate       = "0.2"
	defaultCommissionMaxChangeRate = "0.01"
	defaultMinSelfDelegation       = "1"
)

// GenTxCmd builds the application's gentx command.
// nolint: errcheck
func GenTxCmd(ctx *server.Context, cdc *codec.Codec, mbm sdk.ModuleBasicManager,
	genAccIterator genutil.GenesisAccountsIterator, defaultNodeHome, defaultCLIHome string) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "gentx",
		Short: "Generate a genesis tx carrying a self delegation",
		Args:  cobra.NoArgs,
		Long: fmt.Sprintf(`This command is an alias of the 'tx create-validator' command'.

It creates a genesis piece carrying a self delegation with the
following delegation and commission default parameters:

	delegation amount:           %s
	commission rate:             %s
	commission max rate:         %s
	commission max change rate:  %s
	minimum self delegation:     %s
`, defaultAmount, defaultCommissionRate, defaultCommissionMaxRate, defaultCommissionMaxChangeRate, defaultMinSelfDelegation),
		RunE: func(cmd *cobra.Command, args []string) error {

			config := ctx.Config
			config.SetRoot(viper.GetString(tmcli.HomeFlag))
			nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(ctx.Config)
			if err != nil {
				return err
			}

			// Read --nodeID, if empty take it from priv_validator.json
			if nodeIDString := viper.GetString(flagNodeID); nodeIDString != "" {
				nodeID = nodeIDString
			}

			ip := viper.GetString(flagIP)
			if ip == "" {
				fmt.Fprintf(os.Stderr, "couldn't retrieve an external IP; "+
					"the tx's memo field will be unset")
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

			kb, err := keys.NewKeyBaseFromDir(viper.GetString(flagClientHome))
			if err != nil {
				return err
			}

			name := viper.GetString(client.FlagName)
			key, err := kb.Get(name)
			if err != nil {
				return err
			}

			// Read --pubkey, if empty take it from priv_validator.json
			if valPubKeyString := viper.GetString(flagPubKey); valPubKeyString != "" {
				valPubKey, err = sdk.GetConsPubKeyBech32(valPubKeyString)
				if err != nil {
					return err
				}
			}

			website := viper.GetString(flagWebsite)
			details := viper.GetString(flagDetails)
			identity := viper.GetString(flagIdentity)

			// Set flags for creating gentx
			prepareFlagsForTxCreateValidator(config, nodeID, ip, genDoc.ChainID, valPubKey, website, details, identity)

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

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// XXX: Set the generate-only flag here after the CLI context has
			// been created. This allows the from name/key to be correctly populated.
			//
			// TODO: Consider removing the manual setting of generate-only in
			// favor of a 'gentx' flag in the create-validator command.
			viper.Set(client.FlagGenerateOnly, true)

			// get flag information
			description := staking.NewDescription(
				viper.GetString(flagMoniker),
				viper.GetString(flagIdentity),
				viper.GetString(flagWebsite),
				viper.GetString(flagDetails),
			)
			amounstStr := viper.GetString(flagAmount)
			pkStr := viper.GetString(flagPubKey)
			rateStr := viper.GetString(flagCommissionRate)
			maxRateStr := viper.GetString(flagCommissionMaxRate)
			maxChangeRateStr := viper.GetString(flagCommissionMaxChangeRate)
			msbStr := viper.GetString(flagMinSelfDelegation)
			generateOnly := viper.GetBool(client.FlagGenerateOnly)

			// create a 'create-validator' message
			txBldr, msg, err := cli.BuildCreateValidatorMsg(cliCtx, txBldr,
				amounstStr, pkStr, rateStr, maxRateStr, maxChangeRateStr, msbStr,
				ip, nodeID, description, generateOnly)
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

	ip, _ := server.ExternalIP()

	cmd.Flags().String(tmcli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultCLIHome, "client's home directory")
	cmd.Flags().String(client.FlagName, "", "name of private key with which to sign the gentx")
	cmd.Flags().String(client.FlagOutputDocument, "",
		"write the genesis transaction JSON document to the given file instead of the default location")
	cmd.Flags().String(flagIP, ip, "The node's public IP")
	cmd.Flags().String(flagNodeID, "", "The node's NodeID")
	cmd.Flags().String(flagWebsite, "", "The validator's (optional) website")
	cmd.Flags().String(flagDetails, "", "The validator's (optional) details")
	cmd.Flags().String(flagIdentity, "", "The (optional) identity signature (ex. UPort or Keybase)")
	cmd.Flags().String(flagCommissionRate, "", "The initial commission rate percentage")
	cmd.Flags().String(flagCommissionMaxRate, "", "The maximum commission rate percentage")
	cmd.Flags().String(flagCommissionMaxChangeRate, "", "The maximum commission change rate percentage (per day)")
	cmd.Flags().String(flagMinSelfDelegation, "", "The minimum self delegation required on the validator")
	cmd.Flags().String(flagAmount, "", "Amount of coins to bond")
	cmd.Flags().String(flagPubKey, "", "The Bech32 encoded PubKey of the validator")

	cmd.MarkFlagRequired(client.FlagName)
	return cmd
}

const (
	flagIP                      = "ip"
	flagNodeID                  = "node-id"
	flagWebsite                 = "website"
	flagDetails                 = "details"
	flagPubKey                  = "pubkey"
	flagAmount                  = "amount"
	flagMoniker                 = "moniker"
	flagIdentity                = "identity"
	flagCommissionRate          = "commission-rate"
	flagCommissionMaxRate       = "commission-max-rate"
	flagCommissionMaxChangeRate = "commission-max-change-rate"
	flagMinSelfDelegation       = "min-self-delegation"
)

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
	defer outputFile.Close()
	if err != nil {
		return err
	}
	json, err := cdc.MarshalJSON(tx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(outputFile, "%s\n", json)
	return err
}

func prepareFlagsForTxCreateValidator(
	config *cfg.Config, nodeID, ip, chainID string, valPubKey crypto.PubKey, website, details, identity string,
) {
	viper.Set(tmcli.HomeFlag, viper.GetString(flagClientHome))
	viper.Set(client.FlagChainID, chainID)
	viper.Set(client.FlagFrom, viper.GetString(client.FlagName))
	viper.Set(flagNodeID, nodeID)
	viper.Set(flagIP, ip)
	viper.Set(flagPubKey, sdk.MustBech32ifyConsPub(valPubKey))
	viper.Set(flagMoniker, config.Moniker)
	viper.Set(flagWebsite, website)
	viper.Set(flagDetails, details)
	viper.Set(flagIdentity, identity)

	if config.Moniker == "" {
		viper.Set(flagMoniker, viper.GetString(client.FlagName))
	}
	if viper.GetString(flagAmount) == "" {
		viper.Set(flagAmount, defaultAmount)
	}
	if viper.GetString(flagCommissionRate) == "" {
		viper.Set(flagCommissionRate, defaultCommissionRate)
	}
	if viper.GetString(flagCommissionMaxRate) == "" {
		viper.Set(flagCommissionMaxRate, defaultCommissionMaxRate)
	}
	if viper.GetString(flagCommissionMaxChangeRate) == "" {
		viper.Set(flagCommissionMaxChangeRate, defaultCommissionMaxChangeRate)
	}
	if viper.GetString(flagMinSelfDelegation) == "" {
		viper.Set(flagMinSelfDelegation, defaultMinSelfDelegation)
	}
}

// DONTCOVER
