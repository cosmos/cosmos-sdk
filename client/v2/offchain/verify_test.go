package offchain

import (
	"testing"

	"github.com/stretchr/testify/require"

	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"
	clientcontext "cosmossdk.io/client/v2/context"
	clitx "cosmossdk.io/client/v2/tx"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func Test_Verify(t *testing.T) {
	ctx := clientcontext.Context{
		AddressCodec:          address.NewBech32Codec("cosmos"),
		ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
		Cdc:                   getCodec(),
	}

	tests := []struct {
		name       string
		digest     []byte
		fileFormat string
		wantErr    bool
	}{
		{
			name:       "verify json",
			digest:     []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\", \"app_domain\":\"<appd>\", \"signer\":\"cosmos16877zjk85kwlap3wclpmx34e0xllg2erc7u7m4\", \"data\":\"{\\n\\t\\\"name\\\": \\\"Sarah\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 29\\n}\\n\"}], \"timeout_timestamp\":\"0001-01-01T00:00:00Z\"}, \"auth_info\":{\"signer_infos\":[{\"public_key\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\", \"key\":\"Ahhu3idSSUAQXtDBvBjUlCPWH3od4rXyWgb7L4scSj4m\"}, \"mode_info\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}}}], \"fee\":{}}, \"signatures\":[\"tdXsO5uNqIBFSBKEA1e3Wrcb6ejriP9HwlcBTkU7EUJzuezjg6Rvr1a+Kp6umCAN7MWoBHRT2cmqzDfg6RjaYA==\"]}"),
			fileFormat: "json",
		},
		{
			name:       "wrong signer json",
			digest:     []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\", \"app_domain\":\"<appd>\", \"signer\":\"cosmos1xv9e39mkhhyg5aneu2myj82t7029sv48qu3pgj\", \"data\":\"{\\n\\t\\\"name\\\": \\\"Sarah\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 29\\n}\\n\"}], \"timeout_timestamp\":\"0001-01-01T00:00:00Z\"}, \"auth_info\":{\"signer_infos\":[{\"public_key\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\", \"key\":\"Ahhu3idSSUAQXtDBvBjUlCPWH3od4rXyWgb7L4scSj4m\"}, \"mode_info\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}}}], \"fee\":{}}, \"signatures\":[\"tdXsO5uNqIBFSBKEA1e3Wrcb6ejriP9HwlcBTkU7EUJzuezjg6Rvr1a+Kp6umCAN7MWoBHRT2cmqzDfg6RjaYA==\"]}"),
			fileFormat: "json",
			wantErr:    true,
		},
		{
			name:       "verify text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{app_domain:\"<appd>\" signer:\"cosmos16877zjk85kwlap3wclpmx34e0xllg2erc7u7m4\" data:\"{\\n\\t\\\"name\\\": \\\"Sarah\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 29\\n}\\n\"}} timeout_timestamp:{seconds:-62135596800}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x02\\x18n\\xde'RI@\\x10^\\xd0\\xc1\\xbc\\x18Ԕ#\\xd6\\x1fz\\x1d\\xe2\\xb5\\xf2Z\\x06\\xfb/\\x8b\\x1cJ>&\"}} mode_info:{single:{mode:SIGN_MODE_DIRECT}}} fee:{}} signatures:\"\\xb5\\xd5\\xec;\\x9b\\x8d\\xa8\\x80EH\\x12\\x84\\x03W\\xb7Z\\xb7\\x1b\\xe9\\xe8\\xeb\\x88\\xffG\\xc2W\\x01NE;\\x11Bs\\xb9\\xecヤo\\xafV\\xbe*\\x9e\\xae\\x98 \\r\\xecŨ\\x04tS\\xd9ɪ\\xcc7\\xe0\\xe9\\x18\\xda`\"\n"),
			fileFormat: "text",
		},
		{
			name:       "wrong signer text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{app_domain:\"<appd>\" signer:\"cosmos1xv9e39mkhhyg5aneu2myj82t7029sv48qu3pgj\" data:\"{\\n\\t\\\"name\\\": \\\"Sarah\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 29\\n}\\n\"}} timeout_timestamp:{seconds:-62135596800}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x02\\x18n\\xde'RI@\\x10^\\xd0\\xc1\\xbc\\x18Ԕ#\\xd6\\x1fz\\x1d\\xe2\\xb5\\xf2Z\\x06\\xfb/\\x8b\\x1cJ>&\"}} mode_info:{single:{mode:SIGN_MODE_DIRECT}}} fee:{}} signatures:\"\\xb5\\xd5\\xec;\\x9b\\x8d\\xa8\\x80EH\\x12\\x84\\x03W\\xb7Z\\xb7\\x1b\\xe9\\xe8\\xeb\\x88\\xffG\\xc2W\\x01NE;\\x11Bs\\xb9\\xecヤo\\xafV\\xbe*\\x9e\\xae\\x98 \\r\\xecŨ\\x04tS\\xd9ɪ\\xcc7\\xe0\\xe9\\x18\\xda`\"\n"),
			fileFormat: "text",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Verify(ctx, tt.digest, tt.fileFormat)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_SignVerify(t *testing.T) {
	ac := address.NewBech32Codec("cosmos")

	k := keyring.NewInMemory(getCodec())
	_, err := k.NewAccount("signVerify", mnemonic, "", "m/44'/118'/0'/0/0", hd.Secp256k1)
	require.NoError(t, err)

	autoKeyring, err := keyring.NewAutoCLIKeyring(k, ac)
	require.NoError(t, err)

	ctx := clientcontext.Context{
		AddressCodec:          address.NewBech32Codec("cosmos"),
		ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
		Cdc:                   getCodec(),
		Keyring:               autoKeyring,
	}

	tx, err := Sign(ctx, []byte("Hello World!"), mockClientConn{}, "signVerify", "no-encoding", "direct", "json")
	require.NoError(t, err)

	err = Verify(ctx, []byte(tx), "json")
	require.NoError(t, err)
}

func Test_unmarshal(t *testing.T) {
	txConfig, err := clitx.NewTxConfig(clitx.ConfigOptions{
		AddressCodec:          address.NewBech32Codec("cosmos"),
		Cdc:                   getCodec(),
		ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
		EnabledSignModes:      enabledSignModes,
	})
	require.NoError(t, err)
	tests := []struct {
		name       string
		digest     []byte
		fileFormat string
	}{
		{
			name:       "json test",
			digest:     []byte(`{"body":{"messages":[{"@type":"/offchain.MsgSignArbitraryData", "appDomain":"simd", "signer":"cosmos1x33fy6rusfprkntvjsfregss7rvsvyy4lkwrqu", "data":"{\n\t\"name\": \"John\",\n\t\"surname\": \"Connor\",\n\t\"age\": 15\n}\n"}]}, "authInfo":{"signerInfos":[{"publicKey":{"@type":"/cosmos.crypto.secp256k1.PubKey", "key":"A/Bfsb7grZtysreo48oB1XAXbcgHnEJyhAqzDMgbLlXw"}, "modeInfo":{"single":{"mode":"SIGN_MODE_TEXTUAL"}}}], "fee":{}}, "signatures":["gRufjcmATaJ3hZSiXII3lcsLDJlHM4OhQs3O/QgAK4weQ73kmj30/gw3HwTKxGb4pnVe0iyLXrKRNeSl1O3zSQ=="]}`),
			fileFormat: "json",
		},
		{
			name:       "text test",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{app_domain:\"simd\"  signer:\"cosmos1x33fy6rusfprkntvjsfregss7rvsvyy4lkwrqu\"  data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}}  auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03\\xf0_\\xb1\\xbe\u0B5Br\\xb2\\xb7\\xa8\\xe3\\xca\\x01\\xd5p\\x17m\\xc8\\x07\\x9cBr\\x84\\n\\xb3\\x0c\\xc8\\x1b.U\\xf0\"}}  mode_info:{single:{mode:SIGN_MODE_TEXTUAL}}}  fee:{}}  signatures:\"\\x81\\x1b\\x9f\\x8dɀM\\xa2w\\x85\\x94\\xa2\\\\\\x827\\x95\\xcb\\x0b\\x0c\\x99G3\\x83\\xa1B\\xcd\\xce\\xfd\\x08\\x00+\\x8c\\x1eC\\xbd\\xe4\\x9a=\\xf4\\xfe\\x0c7\\x1f\\x04\\xca\\xc4f\\xf8\\xa6u^\\xd2,\\x8b^\\xb2\\x915\\xe4\\xa5\\xd4\\xed\\xf3I\"\n"),
			fileFormat: "text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshal(tt.fileFormat, tt.digest, txConfig)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
