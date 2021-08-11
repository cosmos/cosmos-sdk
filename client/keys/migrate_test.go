package keys

import (
	"context"
	"fmt"
	"strings"
	"testing"

	design99keyring "github.com/99designs/keyring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
)

type setter interface {
	SetItem(item design99keyring.Item) error
}

func Test_runMigrateCmdLegacyInfo(t *testing.T) {
	const n1 = "cosmos"

	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()
	require := require.New(t)

	// instantiate keyring
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	multi := multisig.NewLegacyAminoPubKey(
		1, []cryptotypes.PubKey{
			priv.PubKey(),
		},
	)
	legacyMultiInfo, err := keyring.NewLegacyMultiInfo(n1, multi)
	require.NoError(err)
	serializedLegacyMultiInfo := keyring.MarshalInfo(legacyMultiInfo)

	// adding LegacyInfo item into keyring
	item := design99keyring.Item{
		Key:         n1,
		Data:        serializedLegacyMultiInfo,
		Description: "SDK kerying version",
	}

	setter, ok := kb.(setter)
	require.True(ok)
	require.NoError(setter.SetItem(item))

	clientCtx := client.Context{}.WithKeyringDir(dir).WithKeyring(kb)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	// run MigrateCommand, it should return no error
	cmd := MigrateCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	mockIn2 := testutil.ApplyMockIODiscardOutErr(cmd)
	mockIn2, mockOut := testutil.ApplyMockIO(cmd)

	mockIn2.Reset("\n12345678\n\n\n\n\n")
	t.Log(mockOut.String())
	assert.NoError(t, cmd.ExecuteContext(ctx))
}

func Test_runMigrateCmdRecord(t *testing.T) {
	const n1 = "cosmos"

	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()
	require := require.New(t)

	// instantiate keyring
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	privKey := cryptotypes.PrivKey(priv)
	k, err := keyring.NewLocalRecord("test record", privKey, privKey.PubKey())
	require.NoError(err)
	serializedRecord, err := encCfg.Codec.Marshal(k)
	require.NoError(err)

	item := design99keyring.Item{
		Key:         n1,
		Data:        serializedRecord,
		Description: "SDK kerying version",
	}

	setter, ok := kb.(setter)
	require.True(ok)
	require.NoError(setter.SetItem(item))

	clientCtx := client.Context{}.WithKeyringDir(dir).WithKeyring(kb)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	// run MigrateCommand, it should return no error
	cmd := MigrateCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	mockIn2 := testutil.ApplyMockIODiscardOutErr(cmd)
	mockIn2, mockOut := testutil.ApplyMockIO(cmd)

	mockIn2.Reset("\n12345678\n\n\n\n\n")
	t.Log(mockOut.String())
	assert.NoError(t, cmd.ExecuteContext(ctx))
}

func Test_runMigrateCmdNoKeys(t *testing.T) {
	kbHome := t.TempDir()
	clientCtx := client.Context{}.WithKeyringDir(kbHome)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	cmd := MigrateCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	cmd.SetArgs([]string{fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest)})

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	mockIn, mockOut := testutil.ApplyMockIO(cmd)

	mockIn.Reset("\n12345678\n\n\n\n\n")
	t.Log(mockOut.String())
	assert.NoError(t, cmd.ExecuteContext(ctx))
}
