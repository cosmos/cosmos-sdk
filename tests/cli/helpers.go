package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/require"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	TotalCoins = sdk.NewCoins(
		sdk.NewCoin(Fee2Denom, sdk.TokensFromConsensusPower(2000000)),
		sdk.NewCoin(FeeDenom, sdk.TokensFromConsensusPower(2000000)),
		sdk.NewCoin(FooDenom, sdk.TokensFromConsensusPower(2000)),
		sdk.NewCoin(Denom, sdk.TokensFromConsensusPower(300).Add(sdk.NewInt(12))), // add coins from inflation
	)

	StartCoins = sdk.NewCoins(
		sdk.NewCoin(Fee2Denom, sdk.TokensFromConsensusPower(1000000)),
		sdk.NewCoin(FeeDenom, sdk.TokensFromConsensusPower(1000000)),
		sdk.NewCoin(FooDenom, sdk.TokensFromConsensusPower(1000)),
		sdk.NewCoin(Denom, sdk.TokensFromConsensusPower(150)),
	)

	VestingCoins = sdk.NewCoins(
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
	cmd := fmt.Sprintf("%s init --overwrite --home=%s %s", f.SimdBinary, f.SimdHome, moniker)
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
	cmd := fmt.Sprintf("%s gentx --name=%s --home=%s --keyring-backend=test --chain-id=%s", f.SimdBinary, name, f.SimdHome, f.ChainID)
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
	cmd := fmt.Sprintf("%s keys delete --keyring-backend=test --home=%s %s", f.SimdBinary,
		f.SimcliHome, name)
	ExecuteWrite(f.T, AddFlags(cmd, append(append(flags, "-y"), "-f")))
}

// KeysAdd is simcli keys add
func (f *Fixtures) KeysAdd(name string, flags ...string) {
	cmd := fmt.Sprintf("%s keys add --keyring-backend=test --home=%s %s", f.SimdBinary,
		f.SimcliHome, name)
	ExecuteWriteCheckErr(f.T, AddFlags(cmd, flags))
}

// KeysAddRecover prepares simcli keys add --recover
func (f *Fixtures) KeysAddRecover(name, mnemonic string, flags ...string) (exitSuccess bool, stdout, stderr string) {
	cmd := fmt.Sprintf("%s keys add --keyring-backend=test --home=%s --recover %s",
		f.SimdBinary, f.SimcliHome, name)
	return ExecuteWriteRetStdStreams(f.T, AddFlags(cmd, flags), mnemonic)
}

// KeysAddRecoverHDPath prepares simcli keys add --recover --account --index
func (f *Fixtures) KeysAddRecoverHDPath(name, mnemonic string, account uint32, index uint32, flags ...string) {
	cmd := fmt.Sprintf("%s keys add --keyring-backend=test --home=%s --recover %s --account %d"+
		" --index %d", f.SimdBinary, f.SimcliHome, name, account, index)
	ExecuteWriteCheckErr(f.T, AddFlags(cmd, flags), mnemonic)
}

// KeysShow is simcli keys show
func (f *Fixtures) KeysShow(name string, flags ...string) keyring.KeyOutput {
	cmd := fmt.Sprintf("%s keys show --keyring-backend=test --home=%s %s --output=json", f.SimdBinary, f.SimcliHome, name)
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
// simcli query txs

// QueryTxs is simcli query txs
func (f *Fixtures) QueryTxs(page, limit int, events ...string) *sdk.SearchTxsResult {
	cmd := fmt.Sprintf("%s query txs --page=%d --limit=%d --events='%s' %v",
		f.SimdBinary, page, limit, buildEventsQueryString(events), f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var result sdk.SearchTxsResult

	err := f.Cdc.UnmarshalJSON([]byte(out), &result)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return &result
}

//utils

func AddFlags(cmd string, flags []string) string {
	for _, f := range flags {
		cmd += " " + f
	}
	return strings.TrimSpace(cmd)
}

func UnmarshalStdTx(t require.TestingT, c codec.JSONMarshaler, s string) (stdTx authtypes.StdTx) {
	require.Nil(t, c.UnmarshalJSON([]byte(s), &stdTx))
	return
}

func buildEventsQueryString(events []string) string {
	return strings.Join(events, "&")
}

func MarshalStdTx(t require.TestingT, c *codec.Codec, stdTx authtypes.StdTx) []byte {
	bz, err := c.MarshalBinaryBare(stdTx)
	require.NoError(t, err)

	return bz
}
