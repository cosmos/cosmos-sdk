package speedtest

import (
	"github.com/spf13/cobra"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AccountCreator func(account *authtypes.BaseAccount, genesisFunds sdk.Coins)

type GenerateTx func() []byte

var (
	numAccounts    = 100
	numTxsPerBlock = 5_000
	numBlocksToRun = 10_000
)

func SpeedTestCmd(ac AccountCreator, gentxer GenerateTx, app servertypes.ABCI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speedtest",
		Short: "execution speedtest",
		Long:  "speedtest is a tool for measuring raw execution TPS of your application",
		RunE: func(cmd *cobra.Command, args []string) error {
			for range numAccounts {

			}
			return nil
		},
	}
	cmd.Flags().IntVar(&numAccounts, "accounts", numAccounts, "number of accounts")
	cmd.Flags().IntVar(&numTxsPerBlock, "txs", numTxsPerBlock, "number of txs")
	cmd.Flags().IntVar(&numBlocksToRun, "blocks", numBlocksToRun, "number of blocks")
	return cmd
}
