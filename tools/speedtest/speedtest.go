package speedtest

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type AccountCreator func() (*authtypes.BaseAccount, sdk.Coins)

type GenerateTx func() []byte

var (
	numAccounts    = 10_000
	numTxsPerBlock = 4_000
	numBlocksToRun = 100
	blockMaxGas    = math.MaxInt64
	blockMaxBytes  = math.MaxInt64
	verifyTxs      = false
)

type GenesisModifier func(codec.Codec, map[string]json.RawMessage)

// NewCmd returns a command that will run an execution test on your application.
// Balances and accounts are automatically added to the chain's state via AccountCreator.
// Your genesis will be modified when instantiating a validator set, so please only modify the genesis by supplying the
// GenesisModifier argument.
// IMPORTANT: When testing the limits of your application, use --verify-txs. If you fill up your blocks passed the allowed
// max gas or max bytes, txs will be ignored, and this can muddy your results. Once you've verified that your configuration
// is processing completely, you may remove the --verify-txs flag to get cleaner results.
func NewCmd(
	createAccount AccountCreator,
	generateTx GenerateTx,
	app servertypes.ABCI,
	cdc codec.Codec,
	defaultGenesis map[string]json.RawMessage,
	chainID string,
	genesisModifiers ...GenesisModifier,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "speedtest",
		Short:   "execution speedtest",
		Long:    "speedtest is a tool for measuring raw execution TPS of your application",
		Example: "speedtest --accounts 20000 --txs 2000 --blocks 10 --block-max-gas 1000000000 --block-max-bytes 1000000000 --verify-txs",
		RunE: func(cmd *cobra.Command, args []string) error {
			accounts := make([]simtestutil.GenesisAccount, 0, numAccounts)
			balances := make([]banktypes.Balance, 0, numAccounts)
			for range numAccounts {
				account, balance := createAccount()
				genesisAcc := simtestutil.GenesisAccount{
					GenesisAccount: account,
					Coins:          balance,
				}
				accounts = append(accounts, genesisAcc)
				balances = append(balances, banktypes.Balance{
					Address: account.Address,
					Coins:   balance,
				})
			}

			vals, err := simtestutil.CreateRandomValidatorSet()
			if err != nil {
				return err
			}

			genAccs := make([]authtypes.GenesisAccount, 0, len(accounts))
			for _, acc := range accounts {
				genAccs = append(genAccs, acc.GenesisAccount)
			}
			genesisState, err := simtestutil.GenesisStateWithValSet(cdc, defaultGenesis, vals, genAccs, balances...)
			if err != nil {
				return err
			}
			for _, genModifier := range genesisModifiers {
				genModifier(cdc, genesisState)
			}

			// init chain must be called to stop deliverState from being nil
			stateBytes, err := cmtjson.MarshalIndent(genesisState, "", " ")
			if err != nil {
				return err
			}

			cp := simtestutil.DefaultConsensusParams
			cp.Block.MaxGas = int64(blockMaxGas)
			cp.Block.MaxBytes = int64(blockMaxBytes)
			_, err = app.InitChain(&types.RequestInitChain{
				ChainId:         chainID,
				Validators:      []types.ValidatorUpdate{},
				ConsensusParams: cp,
				AppStateBytes:   stateBytes,
			})
			if err != nil {
				return fmt.Errorf("failed to InitChain: %w", err)
			}

			// commit genesis changes
			_, err = app.FinalizeBlock(&types.RequestFinalizeBlock{
				Height:             1,
				NextValidatorsHash: vals.Hash(),
			})
			if err != nil {
				return fmt.Errorf("failed to finalize genesis block: %w", err)
			}

			blocks := make([][][]byte, 0, numBlocksToRun)
			for range numBlocksToRun {
				block := make([][]byte, 0, numBlocksToRun)
				for range numTxsPerBlock {
					tx := generateTx()
					block = append(block, tx)
				}
				blocks = append(blocks, block)
			}

			elapsed, err := runBlocks(blocks, app, vals.Proposer.Address, verifyTxs)
			if err != nil {
				return fmt.Errorf("failed to run blocks: %w", err)
			}

			numTxs := numBlocksToRun * numTxsPerBlock
			tps := float64(numTxs) / elapsed.Seconds()
			bps := float64(numBlocksToRun) / elapsed.Seconds()
			cmd.Printf("Finished %d blocks (%d txs) in %s\n", numBlocksToRun, numTxs, elapsed)
			cmd.Printf("TPS: %f\n", tps)
			cmd.Printf("BPS: %f\n", bps)

			return nil
		},
	}
	cmd.Flags().IntVar(&numAccounts, "accounts", numAccounts, "number of accounts")
	cmd.Flags().IntVar(&numTxsPerBlock, "txs", numTxsPerBlock, "number of txs")
	cmd.Flags().IntVar(&numBlocksToRun, "blocks", numBlocksToRun, "number of blocks")
	cmd.Flags().BoolVar(&verifyTxs, "verify-txs", verifyTxs, "verify txs passed. this will loop over all tx results and ensure the code == 0.")
	cmd.Flags().IntVar(&blockMaxGas, "block-max-gas", blockMaxGas, "block max gas")
	cmd.Flags().IntVar(&blockMaxBytes, "block-max-bytes", blockMaxBytes, "block max bytes")
	return cmd
}

func runBlocks(blocks [][][]byte, app servertypes.ABCI, proposer []byte, verify bool) (time.Duration, error) {
	start := time.Now()
	height := int64(1)
	for blockNum, txs := range blocks {
		res, err := app.FinalizeBlock(&types.RequestFinalizeBlock{
			Height:          height,
			Txs:             txs,
			Time:            time.Now(),
			ProposerAddress: proposer,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to finalize block #%d: %w", blockNum, err)
		}
		if verify {
			for _, result := range res.TxResults {
				if result.Code != 0 {
					return 0, fmt.Errorf("tx failed in block %d: code=%d codespace=%s", blockNum, result.Code, result.Codespace)
				}
			}
		}
		_, err = app.Commit()
		if err != nil {
			return 0, fmt.Errorf("failed to commit block #%d: %w", blockNum, err)
		}
		height++
	}
	end := time.Since(start)
	return end, nil
}
