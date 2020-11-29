package keys

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_multiSigKey_Properties(t *testing.T) {
	tmpKey1 := secp256k1.GenPrivKeySecp256k1([]byte("mySecret"))
	pk := multisig.NewPubKeyMultisigThreshold(1, []crypto.PubKey{tmpKey1.PubKey()})
	tmp := keys.NewMultiInfo("myMultisig", pk)

	require.Equal(t, "myMultisig", tmp.GetName())
	require.Equal(t, keys.TypeMulti, tmp.GetType())
	require.Equal(t, "D3923267FA8A3DD367BB768FA8BDC8FF7F89DA3F", tmp.GetPubKey().Address().String())
	require.Equal(t, "cosmos16wfryel63g7axeamw68630wglalcnk3l0zuadc", tmp.GetAddress().String())
}

func Test_showKeysCmd(t *testing.T) {
	cmd := ShowKeysCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "false", cmd.Flag(FlagAddress).DefValue)
	require.Equal(t, "false", cmd.Flag(FlagPublicKey).DefValue)
}

func Test_runShowCmd(t *testing.T) {
	runningUnattended := isRunningUnattended()
	cmd := ShowKeysCmd()
	mockIn, _, _ := tests.ApplyMockIO(cmd)
	require.EqualError(t, runShowCmd(cmd, []string{"invalid"}), "The specified item could not be found in the keyring")
	require.EqualError(t, runShowCmd(cmd, []string{"invalid1", "invalid2"}), "The specified item could not be found in the keyring")

	// Prepare a key base
	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	defer cleanUp()
	viper.Set(flags.FlagHome, kbHome)

	fakeKeyName1 := "runShowCmd_Key1"
	fakeKeyName2 := "runShowCmd_Key2"
	kb, err := keys.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), mockIn)
	require.NoError(t, err)
	defer func() {
		kb.Delete("runShowCmd_Key1", "", false)
		kb.Delete("runShowCmd_Key2", "", false)
	}()
	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	_, err = kb.CreateAccount(fakeKeyName1, tests.TestMnemonic, "", "", sdk.FullFundraiserPath, keys.Secp256k1)
	require.NoError(t, err)

	if runningUnattended {
		mockIn.Reset("testpass1\n")
	}
	_, err = kb.CreateAccount(fakeKeyName2, tests.TestMnemonic, "", "", "m/44'/118'/0'/0/1", keys.Secp256k1)
	require.NoError(t, err)

	// Now try single key
	if runningUnattended {
		mockIn.Reset("testpass1\n")
	}
	require.EqualError(t, runShowCmd(cmd, []string{fakeKeyName1}), "invalid Bech32 prefix encoding provided: ")

	// Now try single key - set bech to acc
	viper.Set(FlagBechPrefix, sdk.PrefixAccount)
	if runningUnattended {
		mockIn.Reset("testpass1\n")
	}
	require.NoError(t, runShowCmd(cmd, []string{fakeKeyName1}))

	// Now try multisig key - set bech to acc
	viper.Set(FlagBechPrefix, sdk.PrefixAccount)
	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	require.EqualError(t, runShowCmd(cmd, []string{fakeKeyName1, fakeKeyName2}), "threshold must be a positive integer")

	// Now try multisig key - set bech to acc + threshold=2
	viper.Set(FlagBechPrefix, sdk.PrefixAccount)
	viper.Set(flagMultiSigThreshold, 2)
	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	err = runShowCmd(cmd, []string{fakeKeyName1, fakeKeyName2})
	require.NoError(t, err)

	// Now try multisig key - set bech to acc + threshold=2
	viper.Set(FlagBechPrefix, "acc")
	viper.Set(FlagDevice, true)
	viper.Set(flagMultiSigThreshold, 2)
	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	err = runShowCmd(cmd, []string{fakeKeyName1, fakeKeyName2})
	require.EqualError(t, err, "the device flag (-d) can only be used for accounts stored in devices")

	viper.Set(FlagBechPrefix, "val")
	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	err = runShowCmd(cmd, []string{fakeKeyName1, fakeKeyName2})
	require.EqualError(t, err, "the device flag (-d) can only be used for accounts")

	viper.Set(FlagPublicKey, true)
	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	err = runShowCmd(cmd, []string{fakeKeyName1, fakeKeyName2})
	require.EqualError(t, err, "the device flag (-d) can only be used for addresses not pubkeys")

	// TODO: Capture stdout and compare
}

func Test_validateMultisigThreshold(t *testing.T) {
	type args struct {
		k     int
		nKeys int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"zeros", args{0, 0}, true},
		{"1-0", args{1, 0}, true},
		{"1-1", args{1, 1}, false},
		{"1-2", args{1, 1}, false},
		{"1-2", args{2, 1}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := validateMultisigThreshold(tt.args.k, tt.args.nKeys); (err != nil) != tt.wantErr {
				t.Errorf("validateMultisigThreshold() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getBechKeyOut(t *testing.T) {
	type args struct {
		bechPrefix string
	}
	tests := []struct {
		name    string
		args    args
		want    bechKeyOutFn
		wantErr bool
	}{
		{"empty", args{""}, nil, true},
		{"wrong", args{"???"}, nil, true},
		{"acc", args{sdk.PrefixAccount}, keys.Bech32KeyOutput, false},
		{"val", args{sdk.PrefixValidator}, keys.Bech32ValKeyOutput, false},
		{"cons", args{sdk.PrefixConsensus}, keys.Bech32ConsKeyOutput, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBechKeyOut(tt.args.bechPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBechKeyOut() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				require.NotNil(t, got)
			}

			// TODO: Still not possible to compare functions
			// Maybe in next release: https://github.com/stretchr/testify/issues/182
			//if &got != &tt.want {
			//	t.Errorf("getBechKeyOut() = %v, want %v", got, tt.want)
			//}
		})
	}
}
