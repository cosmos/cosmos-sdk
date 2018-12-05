package clitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	gaiacli "github.com/cosmos/cosmos-sdk/cmd/gaia/cmd/gaiacli/gaiaclicmd"
	gaiad "github.com/cosmos/cosmos-sdk/cmd/gaia/cmd/gaiad/gaiadcmd"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

const (
	keyFoo   = "foo"
	keyBar   = "bar"
	fooDenom = "fooToken"
	bdenom   = stakeTypes.DefaultBondDenom
)

var fooCoins = sdk.Coins{
	sdk.NewCoin(bdenom, sdk.NewInt(150)),
	sdk.NewCoin(fooDenom, sdk.NewInt(1000)),
}

func TestGaiaCLIMinimumFees(t *testing.T) {
	// t.Parallel()
	f := InitializeFixtures(t)

	// start gaiad server with minimum fees
	go f.GDStart("--minimum_fees=2feeToken")

	tests.WaitForTMStart(f.RPCPort)
	tests.WaitForNextNBlocksTM(1, f.RPCPort)

	fooAddr := f.GCLIKeysAddress(keyFoo)
	barAddr := f.GCLIKeysAddress(keyBar)

	fooAcc := f.GCLIGetAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(bdenom).Int64())

	bytes := f.GCLISend(keyFoo, barAddr, sdk.Coins{sdk.NewCoin(bdenom, sdk.NewInt(10))})
	require.Empty(t, bytes.String())

	f.CleanupDirs()
}

func NewFixtures(t *testing.T) Fixtures {
	gaiadHome, gaiacliHome := getTestingHomeDirs(t.Name())
	gcli := gaiacli.MakeGaiaCLI()
	gd := gaiad.MakeGaiad()
	rpcAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)
	p2pAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)
	return Fixtures{
		T:           t,
		GCLI:        gcli,
		GD:          gd,
		GaiadHome:   gaiadHome,
		GaiaCliHome: gaiacliHome,
		P2PAddr:     p2pAddr,
		RPCAddr:     rpcAddr,
		RPCPort:     port,
	}
}

type Fixtures struct {
	T    *testing.T
	GCLI *cobra.Command
	GD   *cobra.Command

	GaiadHome   string
	GaiaCliHome string

	P2PAddr string
	RPCAddr string
	RPCPort string

	ChainID string
}

func getTestingHomeDirs(name string) (string, string) {
	tmpDir := os.TempDir()
	gaiadHome := fmt.Sprintf("%s%s%s%s.test_gaiad", tmpDir, string(os.PathSeparator), name, string(os.PathSeparator))
	gaiacliHome := fmt.Sprintf("%s%s%s%s.test_gaiacli", tmpDir, string(os.PathSeparator), name, string(os.PathSeparator))
	if _, err := os.Stat(gaiacliHome); os.IsNotExist(err) {
		os.MkdirAll(gaiacliHome, os.ModePerm)
	}
	if _, err := os.Stat(gaiadHome); os.IsNotExist(err) {
		os.MkdirAll(gaiadHome, os.ModePerm)
	}
	return gaiadHome, gaiacliHome
}

func InitializeFixtures(t *testing.T) (f Fixtures) {
	f = NewFixtures(t)

	// Run gaiad unsafe-reset-all and remove the old config, gentx files
	f.GDUnsafeResetAll()
	os.RemoveAll(filepath.Join(f.GaiadHome, "config", "gentx"))

	// Delete old keys if they exist
	f.GCLIKeysDelete(keyFoo, "--force")
	f.GCLIKeysDelete(keyBar, "--force")

	// Add the keys back in
	f.GCLIKeysAdd(keyFoo)
	f.GCLIKeysAdd(keyBar)

	// Initalize gaiad
	f.GDInit()

	// Add genesis account
	f.GDAddGenesisAccount(f.GCLIKeysAddress(keyFoo), fooCoins)

	// Generate bonding transaction
	f.GDGenTx(keyFoo)

	// Collect gentxs and finalize genesis
	f.GDCollectGenTxs()

	return
}

func (f Fixtures) GDStart(flags ...string) {
	buf := f.executeDaemon(flags, "start", f.gdHome(), "--rpc.laddr", f.RPCAddr, "--p2p.laddr", f.P2PAddr)
	if buf.Len() > 0 {
		f.T.Log("START", buf.String())
	}
}

func (f Fixtures) GDUnsafeResetAll(flags ...string) {
	buf := f.executeDaemon(flags, "unsafe-reset-all", f.gdHome())
	if buf.Len() > 0 {
		f.T.Log("UNSAFE_RESET_ALL", buf.String())
	}
}

func (f Fixtures) GDCollectGenTxs(flags ...string) {
	buf := f.executeDaemon(flags, "collect-gentxs", f.gdHome())
	if buf.Len() > 0 {
		f.T.Log("COLLECT_GENTXS", buf.String())
	}
}

func (f Fixtures) GDGenTx(name string, flags ...string) {
	buf := f.executeDaemon(flags, "gentx", "--name", name, "--home-client", f.GaiaCliHome, f.gdHome())
	if buf.Len() > 0 {
		f.T.Log("GENTX", buf.String())
	}
	_, err := os.Stdin.Write([]byte(app.DefaultKeyPass))
	require.NoError(f.T, err)
}

func (f Fixtures) GDAddGenesisAccount(account sdk.AccAddress, coins sdk.Coins, flags ...string) {
	buf := f.executeDaemon(flags, "add-genesis-account", string(account), coins.String(), f.gdHome())
	if buf.Len() > 0 {
		f.T.Log("ADD_GENESIS_ACCOUNT", buf.String())
	}
}

func (f Fixtures) GDInit(flags ...string) {
	buf := f.executeDaemon(flags, "init", "--moniker", keyFoo, f.gdHome())

	// Parse response and set chainID
	var chainID string
	var initRes map[string]json.RawMessage
	err := json.Unmarshal(buf.Bytes(), &initRes)
	require.NoError(f.T, err)
	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(f.T, err)
	f.ChainID = chainID
}

func (f Fixtures) GCLIKeysDelete(name string, flags ...string) {
	_, err := f.executeCLIWError(flags, "keys", "delete", name, f.gcliHome())
	if err.Error() != fmt.Sprintf("Key %s not found", name) {
		require.NoError(f.T, err)
	}
}

func (f Fixtures) GCLIKeysAdd(name string, flags ...string) {
	buf, err := f.executeCLIWError(flags, "keys", "add", name, "--ow", f.gcliHome())
	if err == io.EOF {
		_, err := os.Stdin.Write([]byte(app.DefaultKeyPass))
		require.NoError(f.T, err)
	}
	if buf.Len() > 0 {
		f.T.Log("KEYS ADD", buf.String())
	}
}

func (f Fixtures) GCLIKeysAddress(name string, flags ...string) sdk.AccAddress {
	buf := f.executeCLI(flags, "keys", "show", name, f.gcliHome())
	var ko keys.KeyOutput
	keys.UnmarshalJSON(buf.Bytes(), &ko)
	accAddr, err := sdk.AccAddressFromBech32(ko.Address)
	require.NoError(f.T, err)
	return accAddr
}

func (f Fixtures) GCLIGetAccount(account sdk.AccAddress, flags ...string) auth.BaseAccount {
	buf := f.executeCLI(flags, "query", "account", account.String(), f.gcliHome(), f.node(), f.chainID())
	var initRes map[string]json.RawMessage
	err := json.Unmarshal(buf.Bytes(), &initRes)
	require.NoError(f.T, err, "out %s, err %s", buf, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	err = cdc.UnmarshalJSON(value, &acc)
	require.NoError(f.T, err, "value %v, err %v", string(value), err)
	return acc
}

func (f Fixtures) GCLISend(from string, to sdk.AccAddress, amount sdk.Coins, flags ...string) *bytes.Buffer {
	buf := f.executeCLI(flags,
		"tx", "send", "--from", from,
		"--to", to.String(), "--amount", amount.String(),
		f.gcliHome(), f.node(), f.chainID(),
	)
	return buf
}

func (f Fixtures) gdHome() string {
	return fmt.Sprintf("--home='%s'", f.GaiadHome)
}

func (f Fixtures) gcliHome() string {
	return fmt.Sprintf("--home='%s'", f.GaiaCliHome)
}

func (f Fixtures) node() string {
	return fmt.Sprintf("--node='%s'", f.RPCAddr)
}

func (f Fixtures) chainID() string {
	return fmt.Sprintf("--chain-id=%s", f.ChainID)
}

func (f Fixtures) executeCLI(flags []string, args ...string) *bytes.Buffer {
	buf := new(bytes.Buffer)
	f.GCLI.SetArgs(concat(flags, args))
	f.GCLI.SetOutput(buf)
	err := f.GCLI.Execute()
	require.NoError(f.T, err)
	return buf
}

func (f Fixtures) executeCLIWError(flags []string, args ...string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	f.GCLI.SetArgs(concat(flags, args))
	f.GCLI.SetOutput(buf)
	return buf, f.GCLI.Execute()
}

func (f Fixtures) executeDaemon(flags []string, args ...string) *bytes.Buffer {
	buf := new(bytes.Buffer)
	f.GD.SetArgs(concat(flags, args))
	f.GD.SetOutput(buf)
	err := f.GD.Execute()
	require.NoError(f.T, err)
	return buf
}

func (f Fixtures) CleanupDirs() {
	for _, d := range []string{f.GaiaCliHome, f.GaiadHome} {
		os.RemoveAll(d)
	}
}

func concat(flags []string, args []string) []string {
	for _, f := range flags {
		args = append(args, f)
	}
	return args
}
