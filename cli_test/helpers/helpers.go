package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"os"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/require"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	totalCoins = sdk.NewCoins(
		sdk.NewCoin(Fee2Denom, sdk.TokensFromConsensusPower(2000000)),
		sdk.NewCoin(FeeDenom, sdk.TokensFromConsensusPower(2000000)),
		sdk.NewCoin(FooDenom, sdk.TokensFromConsensusPower(2000)),
		sdk.NewCoin(Denom, sdk.TokensFromConsensusPower(300).Add(sdk.NewInt(12))), // add coins from inflation
	)

	startCoins = sdk.NewCoins(
		sdk.NewCoin(Fee2Denom, sdk.TokensFromConsensusPower(1000000)),
		sdk.NewCoin(FeeDenom, sdk.TokensFromConsensusPower(1000000)),
		sdk.NewCoin(FooDenom, sdk.TokensFromConsensusPower(1000)),
		sdk.NewCoin(Denom, sdk.TokensFromConsensusPower(150)),
	)

	vestingCoins = sdk.NewCoins(
		sdk.NewCoin(FeeDenom, sdk.TokensFromConsensusPower(500000)),
	)
)

//___________________________________________________________________________________
// simd

// UnsafeResetAll is simd unsafe-reset-all
func (f *Fixtures) UnsafeResetAll(flags ...string) {
	cmd := fmt.Sprintf("%s --home=%s unsafe-reset-all", f.SimdBinary, f.SimdHome)
	ExecuteWrite(f.T, AddFlags(cmd, flags))
	err := os.RemoveAll(filepath.Join(f.SimdHome, "config", "gentx"))
	require.NoError(f.T, err)
}

// SDInit is simd init
// NOTE: SDInit sets the ChainID for the Fixtures instance
func (f *Fixtures) SDInit(moniker string, flags ...string) {
	cmd := fmt.Sprintf("%s init -o --home=%s %s", f.SimdBinary, f.SimdHome, moniker)
	_, stderr := tests.ExecuteT(f.T, AddFlags(cmd, flags), clientkeys.DefaultKeyPass)

	var chainID string
	var initRes map[string]json.RawMessage

	err := json.Unmarshal([]byte(stderr), &initRes)
	require.NoError(f.T, err)

	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(f.T, err)

	f.ChainID = chainID
}

// AddGenesisAccount is simd add-genesis-account
func (f *Fixtures) AddGenesisAccount(address sdk.AccAddress, coins sdk.Coins, flags ...string) {
	cmd := fmt.Sprintf("%s add-genesis-account %s %s --home=%s --keyring-backend=test", f.SimdBinary, address, coins, f.SimdHome)
	ExecuteWriteCheckErr(f.T, AddFlags(cmd, flags))
}

// GenTx is simd gentx
func (f *Fixtures) GenTx(name string, flags ...string) {
	cmd := fmt.Sprintf("%s gentx --name=%s --home=%s --home-client=%s --keyring-backend=test", f.SimdBinary, name, f.SimdHome, f.SimcliHome)
	ExecuteWriteCheckErr(f.T, AddFlags(cmd, flags))
}

// CollectGenTxs is simd collect-gentxs
func (f *Fixtures) CollectGenTxs(flags ...string) {
	cmd := fmt.Sprintf("%s collect-gentxs --home=%s", f.SimdBinary, f.SimdHome)
	ExecuteWriteCheckErr(f.T, AddFlags(cmd, flags))
}

// SDStart runs simd start with the appropriate flags and returns a process
func (f *Fixtures) SDStart(flags ...string) *tests.Process {
	cmd := fmt.Sprintf("%s start --home=%s --rpc.laddr=%v --p2p.laddr=%v", f.SimdBinary, f.SimdHome, f.RPCAddr, f.P2PAddr)
	proc := tests.GoExecuteTWithStdout(f.T, AddFlags(cmd, flags))
	tests.WaitForTMStart(f.Port)
	tests.WaitForNextNBlocksTM(1, f.Port)
	return proc
}

// SDTendermint returns the results of simd tendermint [query]
func (f *Fixtures) SDTendermint(query string) string {
	cmd := fmt.Sprintf("%s tendermint %s --home=%s", f.SimdBinary, query, f.SimdHome)
	success, stdout, stderr := ExecuteWriteRetStdStreams(f.T, cmd)
	require.Empty(f.T, stderr)
	require.True(f.T, success)
	return strings.TrimSpace(stdout)
}

// ValidateGenesis runs simd validate-genesis
func (f *Fixtures) ValidateGenesis() {
	cmd := fmt.Sprintf("%s validate-genesis --home=%s", f.SimdBinary, f.SimdHome)
	ExecuteWriteCheckErr(f.T, cmd)
}

//___________________________________________________________________________________
// simcli keys

// KeysDelete is simcli keys delete
func (f *Fixtures) KeysDelete(name string, flags ...string) {
	cmd := fmt.Sprintf("%s keys delete --keyring-backend=test --home=%s %s", f.SimcliBinary,
		f.SimcliHome, name)
	ExecuteWrite(f.T, AddFlags(cmd, append(append(flags, "-y"), "-f")))
}

// KeysAdd is simcli keys add
func (f *Fixtures) KeysAdd(name string, flags ...string) {
	cmd := fmt.Sprintf("%s keys add --keyring-backend=test --home=%s %s", f.SimcliBinary,
		f.SimcliHome, name)
	ExecuteWriteCheckErr(f.T, AddFlags(cmd, flags))
}

// KeysAddRecover prepares simcli keys add --recover
func (f *Fixtures) KeysAddRecover(name, mnemonic string, flags ...string) (exitSuccess bool, stdout, stderr string) {
	cmd := fmt.Sprintf("%s keys add --keyring-backend=test --home=%s --recover %s",
		f.SimcliBinary, f.SimcliHome, name)
	return ExecuteWriteRetStdStreams(f.T, AddFlags(cmd, flags), mnemonic)
}

// KeysAddRecoverHDPath prepares simcli keys add --recover --account --index
func (f *Fixtures) KeysAddRecoverHDPath(name, mnemonic string, account uint32, index uint32, flags ...string) {
	cmd := fmt.Sprintf("%s keys add --keyring-backend=test --home=%s --recover %s --account %d"+
		" --index %d", f.SimcliBinary, f.SimcliHome, name, account, index)
	ExecuteWriteCheckErr(f.T, AddFlags(cmd, flags), mnemonic)
}

// KeysShow is simcli keys show
func (f *Fixtures) KeysShow(name string, flags ...string) keyring.KeyOutput {
	cmd := fmt.Sprintf("%s keys show --keyring-backend=test --home=%s %s", f.SimcliBinary,
		f.SimcliHome, name)
	out, _ := tests.ExecuteT(f.T, AddFlags(cmd, flags), "")
	var ko keyring.KeyOutput
	err := clientkeys.UnmarshalJSON([]byte(out), &ko)
	require.NoError(f.T, err)
	return ko
}

// KeyAddress returns the SDK account address from the key
func (f *Fixtures) KeyAddress(name string) sdk.AccAddress {
	ko := f.KeysShow(name)
	accAddr, err := sdk.AccAddressFromBech32(ko.Address)
	require.NoError(f.T, err)
	return accAddr
}

//___________________________________________________________________________________
// simcli config

// CLIConfig is simcli config
func (f *Fixtures) CLIConfig(key, value string, flags ...string) {
	cmd := fmt.Sprintf("%s config --home=%s %s %s", f.SimcliBinary, f.SimcliHome, key, value)
	ExecuteWriteCheckErr(f.T, AddFlags(cmd, flags))
}

// TxSign is gaiacli tx sign
func (f *Fixtures) TxSign(signer, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx sign %v --keyring-backend=test --from=%s %v", f.SimcliBinary,
		f.Flags(), signer, fileName)
	return ExecuteWriteRetStdStreams(f.T, AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxBroadcast is gaiacli tx broadcast
func (f *Fixtures) TxBroadcast(fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx broadcast %v %v", f.SimcliBinary, f.Flags(), fileName)
	return ExecuteWriteRetStdStreams(f.T, AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxEncode is gaiacli tx encode
func (f *Fixtures) TxEncode(fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx encode %v %v", f.SimcliBinary, f.Flags(), fileName)
	return ExecuteWriteRetStdStreams(f.T, AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

//___________________________________________________________________________________
// gaiacli tx staking

// TxStakingCreateValidator is gaiacli tx staking create-validator
func (f *Fixtures) TxStakingCreateValidator(from, consPubKey string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking create-validator %v --keyring-backend=test --from=%s"+
		" --pubkey=%s", f.SimcliBinary, f.Flags(), from, consPubKey)
	cmd += fmt.Sprintf(" --amount=%v --moniker=%v --commission-rate=%v", amount, from, "0.05")
	cmd += fmt.Sprintf(" --commission-max-rate=%v --commission-max-change-rate=%v", "0.20", "0.10")
	cmd += fmt.Sprintf(" --min-self-delegation=%v", "1")
	return ExecuteWriteRetStdStreams(f.T, AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

func (f *Fixtures) TxStakingEditValidator(from string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking edit-validator %v --keyring-backend=test "+
		"--from=%s", f.SimcliBinary, f.Flags(), from)
	return ExecuteWriteRetStdStreams(f.T, AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

func (f *Fixtures) TxStakingDelegate(validatorOperatorAddress, from string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking delegate %s %v %v --keyring-backend=test "+
		"--from=%s", f.SimcliBinary, validatorOperatorAddress, amount, f.Flags(), from)
	return ExecuteWriteRetStdStreams(f.T, AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

func (f *Fixtures) TxStakingReDelegate(srcOperatorAddress, dstOperatorAddress, from string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking redelegate %s %s %v %v --keyring-backend=test "+
		"--from=%s", f.SimcliBinary, srcOperatorAddress, dstOperatorAddress, amount, f.Flags(), from)
	return ExecuteWriteRetStdStreams(f.T, AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingUnbond is gaiacli tx staking unbond
func (f *Fixtures) TxStakingUnbond(from, shares string, validator sdk.ValAddress, flags ...string) bool {
	cmd := fmt.Sprintf("%s tx staking unbond --keyring-backend=test %s %v --from=%s %v",
		f.SimcliBinary, validator, shares, from, f.Flags())
	return ExecuteWrite(f.T, AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryStakingValidator is gaiacli query staking validator
func (f *Fixtures) QueryStakingValidator(valAddr sdk.ValAddress, flags ...string) staking.Validator {
	cmd := fmt.Sprintf("%s query staking validator %s %v", f.SimcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, AddFlags(cmd, flags), "")
	var validator staking.Validator

	err := f.Cdc.UnmarshalJSON([]byte(out), &validator)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return validator
}

// QueryStakingUnbondingDelegationsFrom is gaiacli query staking unbonding-delegations-from
func (f *Fixtures) QueryStakingUnbondingDelegationsFrom(valAddr sdk.ValAddress, flags ...string) []staking.UnbondingDelegation {
	cmd := fmt.Sprintf("%s query staking unbonding-delegations-from %s %v", f.SimcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, AddFlags(cmd, flags), "")
	var ubds []staking.UnbondingDelegation

	err := f.Cdc.UnmarshalJSON([]byte(out), &ubds)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return ubds
}

// QueryStakingDelegationsTo is gaiacli query staking delegations-to
func (f *Fixtures) QueryStakingDelegationsTo(valAddr sdk.ValAddress, flags ...string) []staking.Delegation {
	cmd := fmt.Sprintf("%s query staking delegations-to %s %v", f.SimcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, AddFlags(cmd, flags), "")
	var delegations []staking.Delegation

	err := f.Cdc.UnmarshalJSON([]byte(out), &delegations)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return delegations
}

// QueryStakingPool is gaiacli query staking pool
func (f *Fixtures) QueryStakingPool(flags ...string) staking.Pool {
	cmd := fmt.Sprintf("%s query staking pool %v", f.SimcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, AddFlags(cmd, flags), "")
	var pool staking.Pool

	err := f.Cdc.UnmarshalJSON([]byte(out), &pool)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return pool
}

// QueryStakingParameters is gaiacli query staking parameters
func (f *Fixtures) QueryStakingParameters(flags ...string) staking.Params {
	cmd := fmt.Sprintf("%s query staking params %v", f.SimcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, AddFlags(cmd, flags), "")
	var params staking.Params

	err := f.Cdc.UnmarshalJSON([]byte(out), &params)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return params
}

//___________________________________________________________________________________
// gaiacli query account

// QueryAccount is gaiacli query account
func (f *Fixtures) QueryAccount(address sdk.AccAddress, flags ...string) auth.BaseAccount {
	cmd := fmt.Sprintf("%s query account %s %v", f.SimcliBinary, address, f.Flags())
	out, _ := tests.ExecuteT(f.T, AddFlags(cmd, flags), "")
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(f.T, err, "out %v, err %v", out, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	err = f.Cdc.UnmarshalJSON(value, &acc)
	require.NoError(f.T, err, "value %v, err %v", string(value), err)
	return acc
}

// QueryBalances executes the bank query balances command for a given address and
// flag set.
func (f *Fixtures) QueryBalances(address sdk.AccAddress, flags ...string) sdk.Coins {
	cmd := fmt.Sprintf("%s query bank balances %s %v", f.SimcliBinary, address, f.Flags())
	out, _ := tests.ExecuteT(f.T, AddFlags(cmd, flags), "")

	var balances sdk.Coins

	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &balances), "out %v\n", out)

	return balances
}
