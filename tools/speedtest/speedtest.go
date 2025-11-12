package speedtest

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type Application interface {
	servertypes.Application
	DefaultGenesis() map[string]json.RawMessage
	Codec() codec.Codec
}

type AccountCreator func() (*authtypes.BaseAccount, sdk.Coins)

type GenerateTx func() []byte

type GenesisModifier func(cdc codec.Codec, genesis map[string]json.RawMessage)

var (
	numAccounts    = 100
	numTxsPerBlock = 5_000
	numBlocksToRun = 10_000
)

func SpeedTestCmd(ac AccountCreator, gentxer GenerateTx, app Application, chainID string, genesisModifiers ...GenesisModifier) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speedtest",
		Short: "execution speedtest",
		Long:  "speedtest is a tool for measuring raw execution TPS of your application",
		RunE: func(cmd *cobra.Command, args []string) error {
			accounts := make([]authtypes.GenesisAccount, 0, numAccounts)
			balances := make([]banktypes.Balance, 0, numAccounts)
			for range numAccounts {
				account, balance := ac()
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

			blocks := make([][][]byte, 0, numBlocksToRun)
			for range numBlocksToRun {
				block := make([][]byte, 0, numBlocksToRun)
				for range numTxsPerBlock {
					tx := gentxer()
					block = append(block, tx)
				}
				blocks = append(blocks, block)
			}

			vals, err := simtestutil.CreateRandomValidatorSet()
			if err != nil {
				return err
			}

			cdc := app.Codec()
			genesisState, err := simtestutil.GenesisStateWithValSet(cdc, app.DefaultGenesis(), vals, accounts, balances...)
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

			// init chain will set the validator set and initialize the genesis accounts
			_, err = app.InitChain(&types.RequestInitChain{
				ChainId:         chainID,
				Validators:      []types.ValidatorUpdate{},
				ConsensusParams: simtestutil.DefaultConsensusParams,
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

			if err := runBlocks(blocks, app, vals.Proposer.Address); err != nil {
				return fmt.Errorf("failed to run blocks: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().IntVar(&numAccounts, "accounts", numAccounts, "number of accounts")
	cmd.Flags().IntVar(&numTxsPerBlock, "txs", numTxsPerBlock, "number of txs")
	cmd.Flags().IntVar(&numBlocksToRun, "blocks", numBlocksToRun, "number of blocks")
	return cmd
}

func runBlocks(blocks [][][]byte, app servertypes.ABCI, proposer []byte) error {

	height := int64(1)
	for blockNum, txs := range blocks {
		_, err := app.FinalizeBlock(&types.RequestFinalizeBlock{
			Height:          height,
			Txs:             txs,
			Time:            time.Now(),
			ProposerAddress: proposer,
		})
		if err != nil {
			return fmt.Errorf("failed to finalize block #%d: %w", blockNum, err)
		}
		_, err = app.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit block #%d: %w", blockNum, err)
		}
		height++
	}
	return nil
}
