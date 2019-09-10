package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

const (
	flagClientHome   = "home-client"
	flagVestingStart = "vesting-start-time"
	flagVestingEnd   = "vesting-end-time"
	flagVestingAmt   = "vesting-amount"
)

// AddGenesisAccountCmd returns add-genesis-account cobra Command.
func AddGenesisAccountCmd(ctx *server.Context, cdc *codec.Codec,
	defaultNodeHome, defaultClientHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-account [address_or_key_name] [coin][,[coin]]",
		Short: "Add genesis account to genesis.json",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			var pubkey crypto.PubKey
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				kb, err := keys.NewKeyBaseFromDir(viper.GetString(flagClientHome))
				if err != nil {
					return err
				}

				info, err := kb.Get(args[0])
				if err != nil {
					return err
				}

				pubkey = info.GetPubKey()
				addr = info.GetAddress()
			}

			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			vestingStart := viper.GetInt64(flagVestingStart)
			vestingEnd := viper.GetInt64(flagVestingEnd)
			vestingAmt, err := sdk.ParseCoins(viper.GetString(flagVestingAmt))
			if err != nil {
				return err
			}

			var genesisAcc exported.GenesisAccount
			baseAcc := types.NewBaseAccount(addr, coins.Sort(), pubkey, 0, 0)
			if !vestingAmt.IsZero() {
				baseVestingAcc := types.NewBaseVestingAccount(
					baseAcc, vestingAmt.Sort(), sdk.Coins{},
					sdk.Coins{}, vestingEnd,
				)

				switch {
				case vestingStart != 0 && vestingEnd != 0:
					genesisAcc = types.NewContinuousVestingAccountRaw(baseVestingAcc, vestingStart)
				case vestingEnd != 0:
					genesisAcc = types.NewDelayedVestingAccountRaw(baseVestingAcc)
				default:
					panic(fmt.Sprintf("invalid genesis vesting account: %+v", baseVestingAcc))
				}
			} else {
				genesisAcc = baseAcc
			}

			if err := genesisAcc.Validate(); err != nil {
				return err
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			// add genesis account to the app state
			var authGenesis types.GenesisState
			cdc.MustUnmarshalJSON(appState[types.ModuleName], &authGenesis)

			for _, acc := range authGenesis.Accounts {
				if acc.GetAddress().Equals(addr) {
					return fmt.Errorf("cannot add account at existing address %s", addr)
				}
			}

			authGenesis.Accounts = append(authGenesis.Accounts, genesisAcc)

			genesisStateBz := cdc.MustMarshalJSON(authGenesis)
			appState[types.ModuleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}

			// export app state
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")
	cmd.Flags().String(flagVestingAmt, "", "amount of coins for vesting accounts")
	cmd.Flags().Uint64(flagVestingStart, 0, "schedule start time (unix epoch) for vesting accounts")
	cmd.Flags().Uint64(flagVestingEnd, 0, "schedule end time (unix epoch) for vesting accounts")
	return cmd
}
