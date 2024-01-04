package keys

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func cleanupKeys(t *testing.T, kb keyring.Keyring, keys ...string) func() {
	t.Helper()
	return func() {
		for _, k := range keys {
			if err := kb.Delete(k); err != nil {
				t.Log("can't delete KB key ", k, err)
			}
		}
	}
}

func Test_runListCmd(t *testing.T) {
	cmd := ListKeysCmd()
	cmd.Flags().AddFlagSet(Commands().PersistentFlags())

	kbHome1 := t.TempDir()
	kbHome2 := t.TempDir()

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	cdc := moduletestutil.MakeTestEncodingConfig().Codec
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome2, mockIn, cdc)
	assert.NilError(t, err)

	clientCtx := client.Context{}.
		WithKeyring(kb).
		WithAddressCodec(addresscodec.NewBech32Codec("cosmos")).
		WithValidatorAddressCodec(addresscodec.NewBech32Codec("cosmosvaloper")).
		WithConsensusAddressCodec(addresscodec.NewBech32Codec("cosmosvalcons"))

	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	path := "" // sdk.GetConfig().GetFullBIP44Path()
	_, err = kb.NewAccount("something", testdata.TestMnemonic, "", path, hd.Secp256k1)
	assert.NilError(t, err)

	t.Cleanup(cleanupKeys(t, kb, "something"))

	testData := []struct {
		name    string
		kbDir   string
		wantErr bool
	}{
		{"keybase: empty", kbHome1, false},
		{"keybase: w/key", kbHome2, false},
	}
	for _, tt := range testData {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cmd.SetArgs([]string{
				fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, tt.kbDir),
			})

			if err := cmd.ExecuteContext(ctx); (err != nil) != tt.wantErr {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_runListCmdWithGivenShowField(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig().Codec
	kbHome1 := t.TempDir()

	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome1, strings.NewReader(""), cdc)
	assert.NilError(t, err)

	clientCtx := client.Context{}.
		WithKeyring(kb).
		WithAddressCodec(addresscodec.NewBech32Codec("cosmos")).
		WithValidatorAddressCodec(addresscodec.NewBech32Codec("cosmosvaloper")).
		WithConsensusAddressCodec(addresscodec.NewBech32Codec("cosmosvalcons"))

	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	path := "" // sdk.GetConfig().GetFullBIP44Path()
	record, err := kb.NewAccount("mock", testdata.TestMnemonic, "", path, hd.Secp256k1)
	assert.NilError(t, err)

	ko, err := MkAccKeyOutput(record, clientCtx.AddressCodec)
	assert.NilError(t, err)

	t.Cleanup(cleanupKeys(t, kb, "mock"))

	testData := []struct {
		providedFlags []string
		expectContain string
		wantErr       bool
		wantErrStr    string
	}{
		{[]string{flagListNames, flagListPubKeys}, "", true, "cannot use both --list-pubkeys and --list-names at once"},
		{[]string{flagListNames, flagListAddresses}, "", true, "cannot use both --list-addresses and --list-names at once"},
		{[]string{flagListAddresses, flagListPubKeys}, "", true, "cannot use both --list-addresses and --list-pubkeys at once"},
		{[]string{flagListNames}, record.Name, false, ""},
		{[]string{flagListPubKeys}, ko.PubKey, false, ""},
		{[]string{flagListAddresses}, ko.Address, false, ""},
	}
	for _, tt := range testData {
		t.Run(fmt.Sprintf("test list with %v", tt.providedFlags), func(t *testing.T) {
			cmd := ListKeysCmd()
			cmd.Flags().AddFlagSet(Commands().PersistentFlags())
			_, mockOut := testutil.ApplyMockIO(cmd)
			args := []string{fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome1)}
			for _, providedFlag := range tt.providedFlags {
				args = append(args, fmt.Sprintf("--%s=true", providedFlag))
			}
			cmd.SetArgs(args)
			err = cmd.ExecuteContext(ctx)
			if tt.wantErr {
				assert.Error(t, err, tt.wantErrStr)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, strings.Contains(mockOut.String(), tt.expectContain))
			}
		})
	}
}

func Test_runListKeyTypeCmd(t *testing.T) {
	cmd := ListKeyTypesCmd()

	cdc := moduletestutil.MakeTestEncodingConfig().Codec
	kbHome := t.TempDir()
	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
	assert.NilError(t, err)

	clientCtx := client.Context{}.
		WithKeyringDir(kbHome).
		WithKeyring(kb)

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{})
	assert.NilError(t, err)
	assert.Assert(t, strings.Contains(out.String(), string(hd.Secp256k1Type)))
}
