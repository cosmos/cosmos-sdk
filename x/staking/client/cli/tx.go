package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	defaultTokens                  = sdk.TokensFromConsensusPower(100)
	defaultAmount                  = defaultTokens.String() + sdk.DefaultBondDenom
	defaultCommissionRate          = "0.1"
	defaultCommissionMaxRate       = "0.2"
	defaultCommissionMaxChangeRate = "0.01"
	defaultMinSelfDelegation       = "1"
)

// NewTxCmd returns a root CLI command handler for all x/staking transaction commands.
func NewTxCmd(clientCtx client.Context) *cobra.Command {
	stakingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Staking transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	stakingTxCmd.AddCommand(flags.PostCommands(
		NewCreateValidatorCmd(clientCtx),
		NewEditValidatorCmd(clientCtx),
		NewDelegateCmd(clientCtx),
		NewRedelegateCmd(clientCtx),
		NewUnbondCmd(clientCtx),
	)...)

	return stakingTxCmd
}

func NewCreateValidatorCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator",
		Short: "create new validator initialized with a self-delegation to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := clientCtx.InitWithInput(cmd.InOrStdin())
			txf := tx.NewFactoryFromDeprecated(clientCtx.Input).WithTxGenerator(clientCtx.TxGenerator).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildCreateValidatorMsg(clientCtx, txf)
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}
	cmd.Flags().AddFlagSet(FlagSetPublicKey())
	cmd.Flags().AddFlagSet(FlagSetAmount())
	cmd.Flags().AddFlagSet(fsDescriptionCreate)
	cmd.Flags().AddFlagSet(FlagSetCommissionCreate())
	cmd.Flags().AddFlagSet(FlagSetMinSelfDelegation())

	cmd.Flags().String(FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", flags.FlagGenerateOnly))
	cmd.Flags().String(FlagNodeID, "", "The node's ID")

	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	_ = cmd.MarkFlagRequired(FlagAmount)
	_ = cmd.MarkFlagRequired(FlagPubKey)
	_ = cmd.MarkFlagRequired(FlagMoniker)

	return cmd
}

func NewEditValidatorCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-validator",
		Short: "edit an existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := clientCtx.InitWithInput(cmd.InOrStdin())

			valAddr := clientCtx.GetFromAddress()
			description := types.NewDescription(
				viper.GetString(FlagMoniker),
				viper.GetString(FlagIdentity),
				viper.GetString(FlagWebsite),
				viper.GetString(FlagSecurityContact),
				viper.GetString(FlagDetails),
			)

			var newRate *sdk.Dec

			commissionRate := viper.GetString(FlagCommissionRate)
			if commissionRate != "" {
				rate, err := sdk.NewDecFromStr(commissionRate)
				if err != nil {
					return fmt.Errorf("invalid new commission rate: %v", err)
				}

				newRate = &rate
			}

			var newMinSelfDelegation *sdk.Int

			minSelfDelegationString := viper.GetString(FlagMinSelfDelegation)
			if minSelfDelegationString != "" {
				msb, ok := sdk.NewIntFromString(minSelfDelegationString)
				if !ok {
					return types.ErrMinSelfDelegationInvalid
				}

				newMinSelfDelegation = &msb
			}

			msg := types.NewMsgEditValidator(sdk.ValAddress(valAddr), description, newRate, newMinSelfDelegation)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			return tx.GenerateOrBroadcastTx(clientCtx, msg)
		},
	}

	cmd.Flags().AddFlagSet(fsDescriptionEdit)
	cmd.Flags().AddFlagSet(fsCommissionUpdate)
	cmd.Flags().AddFlagSet(FlagSetMinSelfDelegation())

	return cmd
}

func NewDelegateCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate [validator-addr] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Delegate liquid tokens to a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Delegate an amount of liquid coins to a validator from your wallet.

Example:
$ %s tx staking delegate cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 1000stake --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := clientCtx.InitWithInput(cmd.InOrStdin())

			amount, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			delAddr := clientCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgDelegate(delAddr, valAddr, amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(clientCtx, msg)
		},
	}

	return cmd
}

func NewRedelegateCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegate [src-validator-addr] [dst-validator-addr] [amount]",
		Short: "Redelegate illiquid tokens from one validator to another",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Redelegate an amount of illiquid staking tokens from one validator to another.

Example:
$ %s tx staking redelegate cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 100stake --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := clientCtx.InitWithInput(cmd.InOrStdin())

			delAddr := clientCtx.GetFromAddress()
			valSrcAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			valDstAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoin(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgBeginRedelegate(delAddr, valSrcAddr, valDstAddr, amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(clientCtx, msg)
		},
	}

	return cmd
}

func NewUnbondCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond [validator-addr] [amount]",
		Short: "Unbond shares from a validator",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unbond an amount of bonded shares from a validator.

Example:
$ %s tx staking unbond cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := clientCtx.InitWithInput(cmd.InOrStdin())

			delAddr := clientCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgUndelegate(delAddr, valAddr, amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(clientCtx, msg)
		},
	}

	return cmd
}

func NewBuildCreateValidatorMsg(clientCtx client.Context, txf tx.Factory) (tx.Factory, sdk.Msg, error) {
	amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return txf, nil, err
	}

	valAddr := clientCtx.GetFromAddress()
	pkStr := viper.GetString(FlagPubKey)

	pk, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, pkStr)
	if err != nil {
		return txf, nil, err
	}

	description := types.NewDescription(
		viper.GetString(FlagMoniker),
		viper.GetString(FlagIdentity),
		viper.GetString(FlagWebsite),
		viper.GetString(FlagSecurityContact),
		viper.GetString(FlagDetails),
	)

	// get the initial validator commission parameters
	rateStr := viper.GetString(FlagCommissionRate)
	maxRateStr := viper.GetString(FlagCommissionMaxRate)
	maxChangeRateStr := viper.GetString(FlagCommissionMaxChangeRate)

	commissionRates, err := buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr)
	if err != nil {
		return txf, nil, err
	}

	// get the initial validator min self delegation
	msbStr := viper.GetString(FlagMinSelfDelegation)

	minSelfDelegation, ok := sdk.NewIntFromString(msbStr)
	if !ok {
		return txf, nil, types.ErrMinSelfDelegationInvalid
	}

	msg := types.NewMsgCreateValidator(
		sdk.ValAddress(valAddr), pk, amount, description, commissionRates, minSelfDelegation,
	)
	if err := msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}

	if viper.GetBool(flags.FlagGenerateOnly) {
		ip := viper.GetString(FlagIP)
		nodeID := viper.GetString(FlagNodeID)

		if nodeID != "" && ip != "" {
			txf = txf.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))
		}
	}

	return txf, msg, nil
}

// Return the flagset, particular flags, and a description of defaults
// this is anticipated to be used with the gen-tx
func CreateValidatorMsgFlagSet(ipDefault string) (fs *flag.FlagSet, defaultsDesc string) {
	fsCreateValidator := flag.NewFlagSet("", flag.ContinueOnError)
	fsCreateValidator.String(FlagIP, ipDefault, "The node's public IP")
	fsCreateValidator.String(FlagNodeID, "", "The node's NodeID")
	fsCreateValidator.String(FlagWebsite, "", "The validator's (optional) website")
	fsCreateValidator.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fsCreateValidator.String(FlagDetails, "", "The validator's (optional) details")
	fsCreateValidator.String(FlagIdentity, "", "The (optional) identity signature (ex. UPort or Keybase)")
	fsCreateValidator.AddFlagSet(FlagSetCommissionCreate())
	fsCreateValidator.AddFlagSet(FlagSetMinSelfDelegation())
	fsCreateValidator.AddFlagSet(FlagSetAmount())
	fsCreateValidator.AddFlagSet(FlagSetPublicKey())

	defaultsDesc = fmt.Sprintf(`
	delegation amount:           %s
	commission rate:             %s
	commission max rate:         %s
	commission max change rate:  %s
	minimum self delegation:     %s
`, defaultAmount, defaultCommissionRate,
		defaultCommissionMaxRate, defaultCommissionMaxChangeRate,
		defaultMinSelfDelegation)

	return fsCreateValidator, defaultsDesc
}

type TxCreateValidatorConfig struct {
	ChainID string
	From    string
	NodeID  string
	Moniker string

	Amount string

	CommissionRate          string
	CommissionMaxRate       string
	CommissionMaxChangeRate string
	MinSelfDelegation       string

	TrustNode bool
	PubKey    string

	IP              string
	Website         string
	SecurityContact string
	Details         string
	Identity        string
}

func PrepareConfigForTxCreateValidator(
	config *cfg.Config, flagSet *flag.FlagSet, nodeID, chainID string, valPubKey crypto.PubKey,
) (TxCreateValidatorConfig, error) {
	c := TxCreateValidatorConfig{}

	ip, err := flagSet.GetString(FlagIP)
	if err != nil {
		return c, err
	}
	if ip == "" {
		_, _ = fmt.Fprintf(os.Stderr, "couldn't retrieve an external IP; "+
			"the tx's memo field will be unset")
	}
	c.IP = ip

	website, err := flagSet.GetString(FlagWebsite)
	if err != nil {
		return c, err
	}
	c.Website = website

	securityContact, err := flagSet.GetString(FlagSecurityContact)
	if err != nil {
		return c, err
	}
	c.SecurityContact = securityContact

	details, err := flagSet.GetString(FlagDetails)
	if err != nil {
		return c, err
	}
	c.SecurityContact = details

	identity, err := flagSet.GetString(FlagIdentity)
	if err != nil {
		return c, err
	}
	c.Identity = identity

	c.ChainID = chainID
	c.From, err = flagSet.GetString(flags.FlagName)
	if err != nil {
		return c, err
	}

	c.Amount, err = flagSet.GetString(FlagAmount)
	if err != nil {
		return c, err
	}

	c.CommissionRate, err = flagSet.GetString(FlagCommissionRate)
	if err != nil {
		return c, err
	}

	c.CommissionMaxRate, err = flagSet.GetString(FlagCommissionMaxRate)
	if err != nil {
		return c, err
	}

	c.CommissionMaxChangeRate, err = flagSet.GetString(FlagCommissionMaxChangeRate)
	if err != nil {
		return c, err
	}

	c.MinSelfDelegation, err = flagSet.GetString(FlagMinSelfDelegation)
	if err != nil {
		return c, err
	}

	c.NodeID = nodeID
	c.TrustNode = true
	c.PubKey = sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey)
	c.Moniker = config.Moniker
	c.Website = website
	c.SecurityContact = securityContact
	c.Details = details
	c.Identity = identity

	if config.Moniker == "" {
		c.Moniker = c.From
	}

	if c.Amount == "" {
		c.Amount = defaultAmount
	}

	if c.CommissionRate == "" {
		c.CommissionRate = defaultCommissionRate
	}

	if c.CommissionMaxRate == "" {
		c.CommissionMaxRate = defaultCommissionMaxRate
	}

	if c.CommissionMaxChangeRate == "" {
		c.CommissionMaxChangeRate = defaultCommissionMaxChangeRate
	}

	if c.MinSelfDelegation == "" {
		c.MinSelfDelegation = defaultMinSelfDelegation
	}

	return c, nil
}

// BuildCreateValidatorMsg makes a new MsgCreateValidator.
func BuildCreateValidatorMsg(clientCtx client.Context, config TxCreateValidatorConfig, txBldr authtypes.TxBuilder, generateOnly bool) (authtypes.TxBuilder, sdk.Msg, error) {
	amounstStr := config.Amount
	amount, err := sdk.ParseCoin(amounstStr)

	if err != nil {
		return txBldr, nil, err
	}

	valAddr := clientCtx.GetFromAddress()
	pkStr := config.PubKey

	pk, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, pkStr)
	if err != nil {
		return txBldr, nil, err
	}

	description := types.NewDescription(
		config.Moniker,
		config.Identity,
		config.Website,
		config.SecurityContact,
		config.Details,
	)

	// get the initial validator commission parameters
	rateStr := config.CommissionRate
	maxRateStr := config.CommissionMaxRate
	maxChangeRateStr := config.CommissionMaxChangeRate
	commissionRates, err := buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr)

	if err != nil {
		return txBldr, nil, err
	}

	// get the initial validator min self delegation
	msbStr := config.MinSelfDelegation
	minSelfDelegation, ok := sdk.NewIntFromString(msbStr)

	if !ok {
		return txBldr, nil, types.ErrMinSelfDelegationInvalid
	}

	msg := types.NewMsgCreateValidator(
		sdk.ValAddress(valAddr), pk, amount, description, commissionRates, minSelfDelegation,
	)

	if generateOnly {
		ip := config.IP
		nodeID := config.NodeID

		if nodeID != "" && ip != "" {
			txBldr = txBldr.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))
		}
	}

	return txBldr, msg, nil
}
