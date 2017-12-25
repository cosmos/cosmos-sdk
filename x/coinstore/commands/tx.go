package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client/commands"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/x/coin"
)

// SendTxCmd is CLI command to send tokens between basecoin accounts
var SendTxCmd = &cobra.Command{
	Use:   "send",
	Short: "send tokens from one account to another",
	RunE:  commands.RequireInit(sendTxCmd),
}

// CreditTxCmd is CLI command to issue credit to one account
var CreditTxCmd = &cobra.Command{
	Use:   "credit",
	Short: "issue credit to one account",
	RunE:  commands.RequireInit(creditTxCmd),
}

//nolint
const (
	FlagTo     = "to"
	FlagAmount = "amount"
	FlagFrom   = "from"
)

func init() {
	flags := SendTxCmd.Flags()
	flags.String(FlagTo, "", "Destination address for the bits")
	flags.String(FlagAmount, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	flags.String(FlagFrom, "", "Address sending coins, if not first signer")

	fs2 := CreditTxCmd.Flags()
	fs2.String(FlagTo, "", "Destination address for the bits")
	fs2.String(FlagAmount, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
}

func sendTxCmd(cmd *cobra.Command, args []string) error {
	tx, err := readSendTxFlags()
	if err != nil {
		return err
	}
	return txcmd.DoTx(tx)
}

func readSendTxFlags() (tx sdk.Tx, err error) {
	// parse to address
	toAddr, err := commands.ParseActor(viper.GetString(FlagTo))
	if err != nil {
		return tx, err
	}

	fromAddr, err := readFromAddr()
	if err != nil {
		return tx, err
	}

	amountCoins, err := coin.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return tx, err
	}

	// craft the inputs and outputs
	tx = coin.NewSendOneTx(fromAddr, toAddr, amountCoins)
	return
}

func creditTxCmd(cmd *cobra.Command, args []string) error {
	tx, err := readCreditTxFlags()
	if err != nil {
		return err
	}
	return txcmd.DoTx(tx)
}

func readCreditTxFlags() (tx sdk.Tx, err error) {
	// parse to address
	toAddr, err := commands.ParseActor(viper.GetString(FlagTo))
	if err != nil {
		return tx, err
	}

	amount, err := coin.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return tx, err
	}

	tx = coin.CreditTx{Debitor: toAddr, Credit: amount}.Wrap()
	return
}

func readFromAddr() (sdk.Actor, error) {
	from := viper.GetString(FlagFrom)
	if from == "" {
		return txcmd.GetSignerAct(), nil
	}
	return commands.ParseActor(from)
}
