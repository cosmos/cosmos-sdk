package offchain

import (
	"testing"

	"github.com/stretchr/testify/require"

	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

const mnemonic = "have embark stumble card pistol fun gauge obtain forget oil awesome lottery unfold corn sure original exist siren pudding spread uphold dwarf goddess card"

func Test_Verify(t *testing.T) {
	fromName := "verifyTest"

	k := keyring.NewInMemory(getCodec())
	_, err := k.NewAccount(fromName, mnemonic, "", "m/44'/118'/0'/0/0", hd.Secp256k1)
	require.NoError(t, err)

	ctx := client.Context{
		Keyring:      k,
		TxConfig:     MakeTestTxConfig(),
		Codec:        getCodec(),
		AddressCodec: address.NewBech32Codec("cosmos"),
	}

	tests := []struct {
		name     string
		fromName string
		digest   []byte
		ctx      client.Context
		wantErr  bool
	}{
		{
			name:     "signMode direct",
			fromName: fromName,
			digest:   []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos15r8vphexk8tnu6gvq0a5dhfs3j06ht9kux78rp\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A6g1szfak0/WeLWfsYOSYBv3UdA8dt/EfYLLHkf1Y5yt\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"Bax6+jG6pGTI+txUqEI3pMfppKNqZu4e7HAH/rYxIkZkGXqJKDgm8Eri3SJD6H9mSDUN2EW1VEiAdp1OZXc6Aw==\"]}"),
			ctx:      ctx,
		},
		{
			name:     "signMode textual",
			fromName: fromName,
			digest:   []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos15r8vphexk8tnu6gvq0a5dhfs3j06ht9kux78rp\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A6g1szfak0/WeLWfsYOSYBv3UdA8dt/EfYLLHkf1Y5yt\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_TEXTUAL\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"sfJ8ymLx+rA4BUlSoWOpO0pLAphvPwif5ztHqJVHdlQ3MKp+N5SXfgAPEEhBRyitS8mi/Y7NBr9TIEpjHFr12A==\"]}"),
			ctx:      ctx,
		},
		{
			name:     "signMode legacyAmino",
			fromName: fromName,
			digest:   []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos15r8vphexk8tnu6gvq0a5dhfs3j06ht9kux78rp\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A6g1szfak0/WeLWfsYOSYBv3UdA8dt/EfYLLHkf1Y5yt\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_LEGACY_AMINO_JSON\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"jn/RsMB79ZLSYPJaXVpZLZDt/gJ+4bNLLfAU0Vspj35+/DXBTj+GFdMvWWrc3emeIbYVgbgwaxghNnlim6JYNA==\"]}"),
			ctx:      ctx,
		},
		{
			name:     "wrong signer",
			fromName: fromName,
			digest:   []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos1rt2xyymh5pvycl8dc00et4mxgr4cpzcdlk8ped\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A6g1szfak0/WeLWfsYOSYBv3UdA8dt/EfYLLHkf1Y5yt\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"Bax6+jG6pGTI+txUqEI3pMfppKNqZu4e7HAH/rYxIkZkGXqJKDgm8Eri3SJD6H9mSDUN2EW1VEiAdp1OZXc6Aw==\"]}"),
			ctx:      ctx,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err = Verify(tt.ctx, tt.digest)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_unmarshall(t *testing.T) {
	tests := []struct {
		name   string
		digest []byte
	}{
		{
			name:   "check",
			digest: []byte(`{"body":{"messages":[{"@type":"/offchain.MsgSignArbitraryData","appDomain":"simd","signer":"cosmos1rt2xyymh5pvycl8dc00et4mxgr4cpzcdlk8ped","data":"{\n\t\"name\": \"John\",\n\t\"surname\": \"Connor\",\n\t\"age\": 15\n}\n"}],"memo":"","timeoutHeight":"0","extensionOptions":[],"nonCriticalExtensionOptions":[]},"authInfo":{"signerInfos":[{"publicKey":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A06XPD6ML7BHSWHsc2u5EtkCXzsCNmhgPaDdNCp5nPF2"},"modeInfo":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gasLimit":"0","payer":"","granter":""},"tip":null},"signatures":["hx8Qo6xZ/Ie0d1TFtiVxSK1rUsRKDEiv1IdcgbkSGYgePYZl6aHJxpSxQDXdIeoZiPeIdrsTkkgjmH4wv2BBdw=="]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshall(tt.digest)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
