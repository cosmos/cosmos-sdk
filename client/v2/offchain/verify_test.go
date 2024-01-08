package offchain

import (
	"testing"

	"github.com/stretchr/testify/require"

	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

const mnemonic = "have embark stumble card pistol fun gauge obtain forget oil awesome lottery unfold corn sure original exist siren pudding spread uphold dwarf goddess card"

func Test_Verify(t *testing.T) {
	ctx := client.Context{
		TxConfig:     MakeTestTxConfig(t),
		Codec:        getCodec(),
		AddressCodec: address.NewBech32Codec("cosmos"),
	}

	tests := []struct {
		name       string
		digest     []byte
		fileFormat string
		ctx        client.Context
		wantErr    bool
	}{
		{
			name:       "signMode direct json",
			digest:     []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos1jcc2frcc4mww897ey4ejphj4x7jaza5w74dmaq\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"Ayw88k8vMspZcgo6qkL8INRzP2HVQJZWu6amPsq+Fg4U\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"RUf2CTYcyeFIijviTAtRN9oqlY7BcaWsQtGvmTVQff0sinh6C1IeL4M2UxakDa1PSVveZyy8gdTsQs3zG43/Kw==\"]}"),
			fileFormat: "json",
			ctx:        ctx,
		},
		{
			name:       "signMode textual json",
			digest:     []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos1jcc2frcc4mww897ey4ejphj4x7jaza5w74dmaq\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"Ayw88k8vMspZcgo6qkL8INRzP2HVQJZWu6amPsq+Fg4U\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_TEXTUAL\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"vAGdBHxFWQUtkHg0zZFgXnYZmUCIIBfdAKVGy7NGaaA8Wyz5E6qpmrDW6929U+L/pKU3u1kZ1JKSYsWgcgPzpw==\"]}"),
			fileFormat: "json",
			ctx:        ctx,
		},
		{
			name:       "signMode legacyAmino json",
			digest:     []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos1jcc2frcc4mww897ey4ejphj4x7jaza5w74dmaq\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"Ayw88k8vMspZcgo6qkL8INRzP2HVQJZWu6amPsq+Fg4U\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_LEGACY_AMINO_JSON\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"83vXwdnwHK4EhA6d8ynRwpee/tE5CBAr3t5tsW/DMLpjxSD1pSmqHpBeB+t19WOK/plBv2ODLMQKZlz/kMdcjg==\"]}"),
			fileFormat: "json",
			ctx:        ctx,
		},
		{
			name:       "wrong signer json",
			digest:     []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos1jcc2frcc4mww897ey4ejphj4x7jaza5w74gmaq\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"Ayw88k8vMspZcgo6qkL8INRzP2HVQJZWu6amPsq+Fg4U\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"RUf2CTYcyeFIijviTAtRN9oqlY7BcaWsQtGvmTVQff0sinh6C1IeL4M2UxakDa1PSVveZyy8gdTsQs3zG43/Kw==\"]}"),
			fileFormat: "json",
			ctx:        ctx,
			wantErr:    true,
		},
		{
			name:       "signMode direct text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{app_domain:\"simd\" signer:\"cosmos1jcc2frcc4mww897ey4ejphj4x7jaza5w74dmaq\" data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03,<\\xf2O/2\\xcaYr\\n:\\xaaB\\xfc \\xd4s?a\\xd5@\\x96V\\xbb\\xa6\\xa6>ʾ\\x16\\x0e\\x14\"}} mode_info:{single:{mode:SIGN_MODE_DIRECT}}} fee:{}} signatures:\"EG\\xf6\\t6\\x1c\\xc9\\xe1H\\x8a;\\xe2L\\x0bQ7\\xda*\\x95\\x8e\\xc1q\\xa5\\xacBѯ\\x995P}\\xfd,\\x8axz\\x0bR\\x1e/\\x836S\\x16\\xa4\\r\\xadOI[\\xdeg,\\xbc\\x81\\xd4\\xecB\\xcd\\xf3\\x1b\\x8d\\xff+\""),
			fileFormat: "text",
			ctx:        ctx,
		},
		{
			name:       "signMode textual text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{app_domain:\"simd\" signer:\"cosmos1jcc2frcc4mww897ey4ejphj4x7jaza5w74dmaq\" data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03,<\\xf2O/2\\xcaYr\\n:\\xaaB\\xfc \\xd4s?a\\xd5@\\x96V\\xbb\\xa6\\xa6>ʾ\\x16\\x0e\\x14\"}} mode_info:{single:{mode:SIGN_MODE_TEXTUAL}}} fee:{}} signatures:\"\\xbc\\x01\\x9d\\x04|EY\\x05-\\x90x4͑`^v\\x19\\x99@\\x88 \\x17\\xdd\\x00\\xa5F˳Fi\\xa0<[,\\xf9\\x13\\xaa\\xa9\\x9a\\xb0\\xd6\\xebݽS\\xe2\\xff\\xa4\\xa57\\xbbY\\x19Ԓ\\x92bŠr\\x03\\xf3\\xa7\""),
			fileFormat: "text",
			ctx:        ctx,
		},
		{
			name:       "signMode legacyAmino text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{app_domain:\"simd\" signer:\"cosmos1jcc2frcc4mww897ey4ejphj4x7jaza5w74dmaq\" data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03,<\\xf2O/2\\xcaYr\\n:\\xaaB\\xfc \\xd4s?a\\xd5@\\x96V\\xbb\\xa6\\xa6>ʾ\\x16\\x0e\\x14\"}} mode_info:{single:{mode:SIGN_MODE_LEGACY_AMINO_JSON}}} fee:{}} signatures:\"\\xf3{\\xd7\\xc1\\xd9\\xf0\\x1c\\xae\\x04\\x84\\x0e\\x9d\\xf3)\\xd1\\u0097\\x9e\\xfe\\xd19\\x08\\x10+\\xde\\xdem\\xb1o\\xc30\\xbac\\xc5 \\xf5\\xa5)\\xaa\\x1e\\x90^\\x07\\xebu\\xf5c\\x8a\\xfe\\x99A\\xbfc\\x83,\\xc4\\nf\\\\\\xff\\x90\\xc7\\\\\\x8e\""),
			fileFormat: "text",
			ctx:        ctx,
		},
		{
			name:       "wrong signer text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{app_domain:\"simd\" signer:\"cosmos1jcc2frcc4mww897ey4ejphj4x7jaza5w74gmaq\" data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03,<\\xf2O/2\\xcaYr\\n:\\xaaB\\xfc \\xd4s?a\\xd5@\\x96V\\xbb\\xa6\\xa6>ʾ\\x16\\x0e\\x14\"}} mode_info:{single:{mode:SIGN_MODE_DIRECT}}} fee:{}} signatures:\"EG\\xf6\\t6\\x1c\\xc9\\xe1H\\x8a;\\xe2L\\x0bQ7\\xda*\\x95\\x8e\\xc1q\\xa5\\xacBѯ\\x995P}\\xfd,\\x8axz\\x0bR\\x1e/\\x836S\\x16\\xa4\\r\\xadOI[\\xdeg,\\xbc\\x81\\xd4\\xecB\\xcd\\xf3\\x1b\\x8d\\xff+\""),
			fileFormat: "text",
			ctx:        ctx,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Verify(tt.ctx, tt.digest, tt.fileFormat)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_SignVerify(t *testing.T) {
	k := keyring.NewInMemory(getCodec())
	_, err := k.NewAccount("signVerify", mnemonic, "", "m/44'/118'/0'/0/0", hd.Secp256k1)
	require.NoError(t, err)

	ctx := client.Context{
		TxConfig:     MakeTestTxConfig(t),
		Codec:        getCodec(),
		AddressCodec: address.NewBech32Codec("cosmos"),
		Keyring:      k,
	}

	tx, err := sign(ctx, "signVerify", "digest", apitxsigning.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	err = verify(ctx, tx)
	require.NoError(t, err)
}

func Test_unmarshal(t *testing.T) {
	tests := []struct {
		name       string
		digest     []byte
		fileFormat string
	}{
		{
			name:       "json test",
			digest:     []byte(`{"body":{"messages":[{"@type":"/offchain.MsgSignArbitraryData","appDomain":"simd","signer":"cosmos1jcc2frcc4mww897ey4ejphj4x7jaza5w74dmaq","data":"{\n\t\"name\": \"John\",\n\t\"surname\": \"Connor\",\n\t\"age\": 15\n}\n"}],"memo":"","timeoutHeight":"0","extensionOptions":[],"nonCriticalExtensionOptions":[]},"authInfo":{"signerInfos":[{"publicKey":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Ayw88k8vMspZcgo6qkL8INRzP2HVQJZWu6amPsq+Fg4U"},"modeInfo":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gasLimit":"0","payer":"","granter":""},"tip":null},"signatures":["RUf2CTYcyeFIijviTAtRN9oqlY7BcaWsQtGvmTVQff0sinh6C1IeL4M2UxakDa1PSVveZyy8gdTsQs3zG43/Kw=="]}`),
			fileFormat: "json",
		},
		{
			name:       "text test",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{app_domain:\"simd\" signer:\"cosmos1jcc2frcc4mww897ey4ejphj4x7jaza5w74dmaq\" data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03,<\\xf2O/2\\xcaYr\\n:\\xaaB\\xfc \\xd4s?a\\xd5@\\x96V\\xbb\\xa6\\xa6>ʾ\\x16\\x0e\\x14\"}} mode_info:{single:{mode:SIGN_MODE_DIRECT}}} fee:{}} signatures:\"EG\\xf6\\t6\\x1c\\xc9\\xe1H\\x8a;\\xe2L\\x0bQ7\\xda*\\x95\\x8e\\xc1q\\xa5\\xacBѯ\\x995P}\\xfd,\\x8axz\\x0bR\\x1e/\\x836S\\x16\\xa4\\r\\xadOI[\\xdeg,\\xbc\\x81\\xd4\\xecB\\xcd\\xf3\\x1b\\x8d\\xff+\""),
			fileFormat: "text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshal(tt.digest, tt.fileFormat)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
