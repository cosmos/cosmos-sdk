package keys

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

func Test_multiSigKey_Properties(t *testing.T) {
	tmpKey1 := secp256k1.GenPrivKeySecp256k1([]byte("mySecret"))

	tmp := multiSigKey{
		name: "myMultisig",
		key:  tmpKey1.PubKey(),
	}

	assert.Equal(t, "myMultisig", tmp.GetName())
	assert.Equal(t, keys.TypeLocal, tmp.GetType())
	assert.Equal(t, "015ABFFB09DB738A45745A91E8C401423ECE4016", tmp.GetPubKey().Address().String())
	assert.Equal(t, "cosmos1q9dtl7cfmdec53t5t2g733qpgglvusqk6xdntl", tmp.GetAddress().String())
}

func Test_showKeysCmd(t *testing.T) {
	cmd := showKeysCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "false", cmd.Flag(FlagAddress).DefValue)
	assert.Equal(t, "false", cmd.Flag(FlagPublicKey).DefValue)
}

func Test_runShowCmd(t *testing.T) {
	cmd := showKeysCmd()

	err := runShowCmd(cmd, []string{"invalid"})
	assert.EqualError(t, err, "Key invalid not found")

	err = runShowCmd(cmd, []string{"invalid1", "invalid2"})
	assert.EqualError(t, err, "Key invalid1 not found")

	// Prepare a key base
	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	defer cleanUp()
	viper.Set(cli.HomeFlag, kbHome)

	fakeKeyName1 := "runShowCmd_Key1"
	fakeKeyName2 := "runShowCmd_Key2"
	kb, err := NewKeyBaseFromHomeFlag()
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName1, tests.TestMnemonic, "", "", 0, 0)
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName2, tests.TestMnemonic, "", "", 0, 1)
	assert.NoError(t, err)

	// Now try single key
	err = runShowCmd(cmd, []string{fakeKeyName1})
	assert.EqualError(t, err, "invalid Bech32 prefix encoding provided: ")

	// Now try single key - set bech to acc
	viper.Set(FlagBechPrefix, "acc")
	err = runShowCmd(cmd, []string{fakeKeyName1})
	assert.NoError(t, err)

	// Now try multisig key - set bech to acc
	viper.Set(FlagBechPrefix, "acc")
	err = runShowCmd(cmd, []string{fakeKeyName1, fakeKeyName2})
	assert.EqualError(t, err, "threshold must be a positive integer")

	// Now try multisig key - set bech to acc + threshold=2
	viper.Set(FlagBechPrefix, "acc")
	viper.Set(flagMultiSigThreshold, 2)
	err = runShowCmd(cmd, []string{fakeKeyName1, fakeKeyName2})
	assert.NoError(t, err)

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
		{"acc", args{"acc"}, Bech32KeyOutput, false},
		{"val", args{"val"}, Bech32ValKeyOutput, false},
		{"cons", args{"cons"}, Bech32ConsKeyOutput, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBechKeyOut(tt.args.bechPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBechKeyOut() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.NotNil(t, got)
			}

			// TODO: Still not possible to compare functions
			// Maybe in next release: https://github.com/stretchr/testify/issues/182
			//if &got != &tt.want {
			//	t.Errorf("getBechKeyOut() = %v, want %v", got, tt.want)
			//}
		})
	}
}
