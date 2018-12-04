package clitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	gaiacli "github.com/cosmos/cosmos-sdk/cmd/gaia/cmd/gaiacli/gaiaclicmd"
	gaiad "github.com/cosmos/cosmos-sdk/cmd/gaia/cmd/gaiad/gaiadcmd"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

const (
	keyFoo = "foo"
	keyBar = "bar"
)

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
	return gaiadHome, gaiacliHome
}

func initializeFixtures(t *testing.T) (chainID, servAddr, port, gaiadHome, gaiacliHome, p2pAddr string) {
	f := NewFixtures(t)

	// Run gaiad unsafe-reset-all and remove the old config, gentx files
	f.GDUnsafeResetAll()
	os.RemoveAll(filepath.Join(gaiadHome, "config", "gentx"))

	// Delete old keys if they exist
	f.GCLIKeysDelete(keyFoo, "--force")
	f.GCLIKeysDelete(keyBar, "--force")

	// Add the keys back in
	f.GCLIKeysAdd(keyFoo)
	f.GCLIKeysAdd(keyBar)

	fooAddr := f.GCLIKeysAddress(keyFoo)

	// Initalize gaiad
	f.GDInit()

	executeWriteCheckErr(t, fmt.Sprintf(
		"gaiad add-genesis-account %s 150%s,1000fooToken --home=%s", fooAddr, stakeTypes.DefaultBondDenom, gaiadHome))
	executeWrite(t, fmt.Sprintf("cat %s%sconfig%sgenesis.json", gaiadHome, string(os.PathSeparator), string(os.PathSeparator)))
	executeWriteCheckErr(t, fmt.Sprintf(
		"gaiad gentx --name=foo --home=%s --home-client=%s", gaiadHome, gaiacliHome), app.DefaultKeyPass)
	executeWriteCheckErr(t, fmt.Sprintf("gaiad collect-gentxs --home=%s", gaiadHome), app.DefaultKeyPass)
	// get a free port, also setup some common flags

	return
}

func (f Fixtures) GDUnsafeResetAll(flags ...string) {
	buf := new(bytes.Buffer)
	f.GD.SetOutput(buf)
	run := []string{
		"unsafe-reset-all",
		f.gdHomeFlag(),
	}
	for _, fl := range flags {
		run = append(run, fl)
	}
	f.GD.SetArgs(run)
	err := f.GD.Execute()
	require.NoError(f.T, err)
	if buf.Len() > 0 {
		f.T.Log(buf.String())
	}
}

func (f Fixtures) GDInit(flags ...string) {
	buf := new(bytes.Buffer)
	f.GD.SetOutput(buf)
	run := []string{
		"init",
		"--moniker=foo",
		f.gdHomeFlag(),
	}
	for _, fl := range flags {
		run = append(run, fl)
	}
	f.GD.SetArgs(run)
	err := f.GD.Execute()
	require.NoError(f.T, err)
	var chainID string
	var initRes map[string]json.RawMessage
	err = json.Unmarshal(buf.Bytes(), &initRes)
	require.NoError(f.T, err)

	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(f.T, err)
	f.ChainID = chainID
}

func (f Fixtures) GCLIKeysDelete(name string, flags ...string) {
	buf := new(bytes.Buffer)
	f.GCLI.SetOutput(buf)
	run := []string{
		"keys",
		"delete",
		name,
		f.gcliHomeFlag(),
	}
	for _, fl := range flags {
		run = append(run, fl)
	}
	f.GCLI.SetArgs(run)
	err := f.GCLI.Execute()
	require.NoError(f.T, err)
	if buf.Len() > 0 {
		f.T.Log(buf.String())
	}
}

func (f Fixtures) GCLIKeysAdd(name string, flags ...string) {
	buf := new(bytes.Buffer)
	f.GCLI.SetOutput(buf)
	run := []string{
		"keys",
		"add",
		name,
		f.gcliHomeFlag(),
	}
	for _, fl := range flags {
		run = append(run, fl)
	}
	f.GCLI.SetArgs(run)
	f.GCLI.OutOrStderr()
	err := f.GCLI.Execute()
	require.NoError(f.T, err)
	if buf.Len() > 0 {
		f.T.Log(buf.String())
	}
	_, err = os.Stdin.Write([]byte(app.DefaultKeyPass))
	require.NoError(f.T, err)
}

func (f Fixtures) GCLIKeysAddress(name string, flags ...string) sdk.AccAddress {
	buf := new(bytes.Buffer)
	f.GCLI.SetOutput(buf)
	run := []string{
		"keys",
		"show",
		name,
		f.gcliHomeFlag(),
	}
	for _, fl := range flags {
		run = append(run, fl)
	}
	f.GCLI.SetArgs(run)
	f.GCLI.OutOrStderr()
	err := f.GCLI.Execute()
	require.NoError(f.T, err)
	var ko keys.KeyOutput
	keys.UnmarshalJSON(buf.Bytes(), &ko)
	accAddr, err := sdk.AccAddressFromBech32(ko.Address)
	require.NoError(f.T, err)
	return accAddr
}

func (f Fixtures) gdHomeFlag() string {
	return fmt.Sprintf("--home='%s'", f.GaiadHome)
}

func (f Fixtures) gcliHomeFlag() string {
	return fmt.Sprintf("--home='%s'", f.GaiaCliHome)
}
