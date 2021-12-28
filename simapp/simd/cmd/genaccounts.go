package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const (
	FlagVestingStart         = "vesting-start-time"
	FlagVestingEnd           = "vesting-end-time"
	FlagVestingAmt           = "vesting-amount"
	FlagVestingPeriodsNumber = "vesting-periods-number"
	FlagVestingPeriodsAmts   = "vesting-periods-amounts"
)

// AddGenesisAccountCmd returns add-genesis-account cobra Command.
func AddGenesisAccountCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-account [address_or_key_name] [coin][,[coin]]",
		Short: "Add a genesis account to genesis.json",
		Long: `Add a genesis account to genesis.json. The provided account must specify
the account address or key name and a list of initial coins. If a key name is given,
the address will be looked up in the local Keybase. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			var kr keyring.Keyring
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
				if keyringBackend != "" && clientCtx.Keyring == nil {
					var err error
					kr, err = keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf)
					if err != nil {
						return err
					}
				} else {
					kr = clientCtx.Keyring
				}

				info, err := kr.Key(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keyring: %w", err)
				}
				addr = info.GetAddress()
			}

			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse coins: %w", err)
			}

			vestingStart, _ := cmd.Flags().GetInt64(FlagVestingStart)
			vestingEnd, _ := cmd.Flags().GetInt64(FlagVestingEnd)
			vestingAmtStr, _ := cmd.Flags().GetString(FlagVestingAmt)
			vestingPeriodsNumber, _ := cmd.Flags().GetInt64(FlagVestingPeriodsNumber)
			vestingPeriodsAmts, _ := cmd.Flags().GetString(FlagVestingPeriodsAmts)

			vestingAmt, err := sdk.ParseCoinsNormalized(vestingAmtStr)
			if err != nil {
				return fmt.Errorf("failed to parse vesting amount: %w", err)
			}

			// create concrete account type based on input parameters
			var genAccount authtypes.GenesisAccount

			balances := banktypes.Balance{Address: addr.String(), Coins: coins.Sort()}
			baseAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)

			if !vestingAmt.IsZero() {
				baseVestingAccount := authvesting.NewBaseVestingAccount(baseAccount, vestingAmt.Sort(), vestingEnd)

				if (balances.Coins.IsZero() && !baseVestingAccount.OriginalVesting.IsZero()) ||
					baseVestingAccount.OriginalVesting.IsAnyGT(balances.Coins) {
					return errors.New("vesting amount cannot be greater than total amount")
				}

				switch {

				case vestingStart > 0 && vestingEnd > 0 && (vestingPeriodsNumber > 0 || vestingPeriodsAmts != ""):
					{
						if vestingPeriodsNumber > 0 && vestingPeriodsAmts != "" {
							return fmt.Errorf("the flags %s and %s can't be used at the same time", FlagVestingPeriodsNumber, FlagVestingPeriodsAmts)
						}

						periods := make(authvesting.Periods, 0)

						if vestingPeriodsNumber > 0 {

							if vestingPeriodsNumber < 1 {
								return fmt.Errorf("parameter %s must be >=2 since at least 2 periods start and and are required", FlagVestingPeriodsNumber)
							}

							vestingDenom := vestingAmt.GetDenomByIndex(0)
							periodLength := (vestingEnd - vestingStart) / (vestingPeriodsNumber - 1)
							periodAmount := vestingAmt.AmountOf(vestingDenom).Quo(sdk.NewInt(vestingPeriodsNumber))
							periodCoins := sdk.NewCoins(sdk.NewCoin(vestingDenom, periodAmount))

							for i := 0; i < int(vestingPeriodsNumber); i++ {
								periods = append(periods, authvesting.Period{
									Length: periodLength,
									Amount: periodCoins,
								})
							}
							if len(periods) > 0 {
								// execute the first period at the start date
								periods[0].Length = 0
							}
						}

						if vestingPeriodsAmts != "" {
							amtSum := make(sdk.Coins, 0)
							periodLengthSum := 0

							vestingPeriodsAmtsStrings := strings.Split(vestingPeriodsAmts, ",")
							for _, vestingPeriodString := range vestingPeriodsAmtsStrings {
								vestingPeriodPair := strings.Split(vestingPeriodString, "|")
								if len(vestingPeriodPair) != 2 {
									return fmt.Errorf("vestingPeriodPair %s is invalid", vestingPeriodString)
								}
								epochString := vestingPeriodPair[0]
								periodLength, err := strconv.Atoi(epochString)
								if err != nil || periodLength <= 0 {
									return fmt.Errorf("invalid epoch in periods: %w", err)
								}

								coinAmountString := vestingPeriodPair[1]
								periodCoins, err := sdk.ParseCoinsNormalized(coinAmountString)
								if err != nil {
									return fmt.Errorf("failed to parse vesting amount: %w", err)
								}

								periods = append(periods, authvesting.Period{
									Length: int64(periodLength),
									Amount: periodCoins,
								})

								amtSum = amtSum.Add(periodCoins...)
								periodLengthSum += periodLength
							}
						}

						genAccount = authvesting.NewPeriodicVestingAccountRaw(baseVestingAccount, vestingStart, periods)
					}

				case vestingStart != 0 && vestingEnd != 0:
					genAccount = authvesting.NewContinuousVestingAccountRaw(baseVestingAccount, vestingStart)

				case vestingEnd != 0:
					genAccount = authvesting.NewDelayedVestingAccountRaw(baseVestingAccount)

				default:
					return errors.New("invalid vesting parameters; must supply start and end time or end time")
				}
			} else {
				genAccount = baseAccount
			}

			if err := genAccount.Validate(); err != nil {
				return fmt.Errorf("failed to validate new genesis account: %w", err)
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			authGenState := authtypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)

			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				return fmt.Errorf("failed to get accounts from any: %w", err)
			}

			if accs.Contains(addr) {
				return fmt.Errorf("cannot add account at existing address %s", addr)
			}

			// Add the new account to the set of genesis accounts and sanitize the
			// accounts afterwards.
			accs = append(accs, genAccount)
			accs = authtypes.SanitizeGenesisAccounts(accs)

			genAccs, err := authtypes.PackAccounts(accs)
			if err != nil {
				return fmt.Errorf("failed to convert accounts into any's: %w", err)
			}
			authGenState.Accounts = genAccs

			authGenStateBz, err := clientCtx.Codec.MarshalJSON(&authGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}

			appState[authtypes.ModuleName] = authGenStateBz

			bankGenState := banktypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)
			bankGenState.Balances = append(bankGenState.Balances, balances)
			bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)
			bankGenState.Supply = bankGenState.Supply.Add(balances.Coins...)

			bankGenStateBz, err := clientCtx.Codec.MarshalJSON(bankGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
			}

			appState[banktypes.ModuleName] = bankGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(FlagVestingAmt, "", "amount of coins for vesting accounts")
	cmd.Flags().Int64(FlagVestingStart, 0, "schedule start time (unix epoch) for vesting accounts")
	cmd.Flags().Int64(FlagVestingEnd, 0, "schedule end time (unix epoch) for vesting accounts")
	cmd.Flags().Int64(FlagVestingPeriodsNumber, 0, "number of periods for vesting before the start and end time")
	cmd.Flags().String(FlagVestingPeriodsAmts, "", "the array of the \"epoch|amountDenom,epoch|amountDenom\", values for the periodic for vesting")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
