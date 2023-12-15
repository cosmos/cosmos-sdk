package offchain

import (
	"testing"

	"github.com/stretchr/testify/require"

	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/address"
)

const mnemonic = "have embark stumble card pistol fun gauge obtain forget oil awesome lottery unfold corn sure original exist siren pudding spread uphold dwarf goddess card"

func Test_Verify(t *testing.T) {
	ctx := client.Context{
		TxConfig:     MakeTestTxConfig(),
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
			digest:     []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos15r8vphexk8tnu6gvq0a5dhfs3j06ht9kux78rp\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A6g1szfak0/WeLWfsYOSYBv3UdA8dt/EfYLLHkf1Y5yt\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"Bax6+jG6pGTI+txUqEI3pMfppKNqZu4e7HAH/rYxIkZkGXqJKDgm8Eri3SJD6H9mSDUN2EW1VEiAdp1OZXc6Aw==\"]}"),
			fileFormat: "json",
			ctx:        ctx,
		},
		{
			name:       "signMode textual json",
			digest:     []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos15r8vphexk8tnu6gvq0a5dhfs3j06ht9kux78rp\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A6g1szfak0/WeLWfsYOSYBv3UdA8dt/EfYLLHkf1Y5yt\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_TEXTUAL\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"sfJ8ymLx+rA4BUlSoWOpO0pLAphvPwif5ztHqJVHdlQ3MKp+N5SXfgAPEEhBRyitS8mi/Y7NBr9TIEpjHFr12A==\"]}"),
			fileFormat: "json",
			ctx:        ctx,
		},
		{
			name:       "signMode legacyAmino json",
			digest:     []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos15r8vphexk8tnu6gvq0a5dhfs3j06ht9kux78rp\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A6g1szfak0/WeLWfsYOSYBv3UdA8dt/EfYLLHkf1Y5yt\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_LEGACY_AMINO_JSON\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"jn/RsMB79ZLSYPJaXVpZLZDt/gJ+4bNLLfAU0Vspj35+/DXBTj+GFdMvWWrc3emeIbYVgbgwaxghNnlim6JYNA==\"]}"),
			fileFormat: "json",
			ctx:        ctx,
		},
		{
			name:       "wrong signer json",
			digest:     []byte("{\"body\":{\"messages\":[{\"@type\":\"/offchain.MsgSignArbitraryData\",\"appDomain\":\"simd\",\"signer\":\"cosmos1rt2xyymh5pvycl8dc00et4mxgr4cpzcdlk8ped\",\"data\":\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}],\"memo\":\"\",\"timeoutHeight\":\"0\",\"extensionOptions\":[],\"nonCriticalExtensionOptions\":[]},\"authInfo\":{\"signerInfos\":[{\"publicKey\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A6g1szfak0/WeLWfsYOSYBv3UdA8dt/EfYLLHkf1Y5yt\"},\"modeInfo\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}},\"sequence\":\"0\"}],\"fee\":{\"amount\":[],\"gasLimit\":\"0\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"Bax6+jG6pGTI+txUqEI3pMfppKNqZu4e7HAH/rYxIkZkGXqJKDgm8Eri3SJD6H9mSDUN2EW1VEiAdp1OZXc6Aw==\"]}"),
			fileFormat: "json",
			ctx:        ctx,
			wantErr:    true,
		},
		{
			name:       "signMode direct text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{appDomain:\"simd\" signer:\"cosmos15r8vphexk8tnu6gvq0a5dhfs3j06ht9kux78rp\" data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03\\xa85\\xb37ړO\\xd6x\\xb5\\x9f\\xb1\\x83\\x92`\\x1b\\xf7Q\\xd0<v\\xdf\\xc4}\\x82\\xcb\\x1eG\\xf5c\\x9c\\xad\"}} mode_info:{single:{mode:SIGN_MODE_DIRECT}}} fee:{}} signatures:\"\\x05\\xacz\\xfa1\\xba\\xa4d\\xc8\\xfa\\xdcT\\xa8B7\\xa4\\xc7餣jf\\xee\\x1e\\xecp\\x07\\xfe\\xb61\\\"Fd\\x19z\\x89(8&\\xf0J\\xe2\\xdd\\\"C\\xe8\\x7ffH5\\r\\xd8E\\xb5TH\\x80v\\x9dNew:\\x03\""),
			fileFormat: "text",
			ctx:        ctx,
		},
		{
			name:       "signMode textual text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{appDomain:\"simd\" signer:\"cosmos15r8vphexk8tnu6gvq0a5dhfs3j06ht9kux78rp\" data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03\\xa85\\xb37ړO\\xd6x\\xb5\\x9f\\xb1\\x83\\x92`\\x1b\\xf7Q\\xd0<v\\xdf\\xc4}\\x82\\xcb\\x1eG\\xf5c\\x9c\\xad\"}} mode_info:{single:{mode:SIGN_MODE_TEXTUAL}}} fee:{}} signatures:\"\\xb1\\xf2|\\xcab\\xf1\\xfa\\xb08\\x05IR\\xa1c\\xa9;JK\\x02\\x98o?\\x08\\x9f\\xe7;G\\xa8\\x95GvT70\\xaa~7\\x94\\x97~\\x00\\x0f\\x10HAG(\\xadKɢ\\xfd\\x8e\\xcd\\x06\\xbfS Jc\\x1cZ\\xf5\\xd8\""),
			fileFormat: "text",
			ctx:        ctx,
		},
		{
			name:       "signMode legacyAmino text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{appDomain:\"simd\" signer:\"cosmos15r8vphexk8tnu6gvq0a5dhfs3j06ht9kux78rp\" data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03\\xa85\\xb37ړO\\xd6x\\xb5\\x9f\\xb1\\x83\\x92`\\x1b\\xf7Q\\xd0<v\\xdf\\xc4}\\x82\\xcb\\x1eG\\xf5c\\x9c\\xad\"}} mode_info:{single:{mode:SIGN_MODE_LEGACY_AMINO_JSON}}} fee:{}} signatures:\"\\x8e\\x7fѰ\\xc0{\\xf5\\x92\\xd2`\\xf2Z]ZY-\\x90\\xed\\xfe\\x02~\\xe1\\xb3K-\\xf0\\x14\\xd1[)\\x8f~~\\xfc5\\xc1N?\\x86\\x15\\xd3/Yj\\xdc\\xdd\\xe9\\x9e!\\xb6\\x15\\x81\\xb80k\\x18!6yb\\x9b\\xa2X4\""),
			fileFormat: "text",
			ctx:        ctx,
		},
		{
			name:       "wrong signer text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{appDomain:\"simd\" signer:\"cosmos1rt2xyymh5pvycl8dc00et4mxgr4cpzcdlk8ped\" data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03\\xa85\\xb37ړO\\xd6x\\xb5\\x9f\\xb1\\x83\\x92`\\x1b\\xf7Q\\xd0<v\\xdf\\xc4}\\x82\\xcb\\x1eG\\xf5c\\x9c\\xad\"}} mode_info:{single:{mode:SIGN_MODE_DIRECT}}} fee:{}} signatures:\"\\x05\\xacz\\xfa1\\xba\\xa4d\\xc8\\xfa\\xdcT\\xa8B7\\xa4\\xc7餣jf\\xee\\x1e\\xecp\\x07\\xfe\\xb61\\\"Fd\\x19z\\x89(8&\\xf0J\\xe2\\xdd\\\"C\\xe8\\x7ffH5\\r\\xd8E\\xb5TH\\x80v\\x9dNew:\\x03\""),
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

func Test_unmarshal(t *testing.T) {
	tests := []struct {
		name       string
		digest     []byte
		fileFormat string
	}{
		{
			name:       "check",
			digest:     []byte(`{"body":{"messages":[{"@type":"/offchain.MsgSignArbitraryData","appDomain":"simd","signer":"cosmos1rt2xyymh5pvycl8dc00et4mxgr4cpzcdlk8ped","data":"{\n\t\"name\": \"John\",\n\t\"surname\": \"Connor\",\n\t\"age\": 15\n}\n"}],"memo":"","timeoutHeight":"0","extensionOptions":[],"nonCriticalExtensionOptions":[]},"authInfo":{"signerInfos":[{"publicKey":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A06XPD6ML7BHSWHsc2u5EtkCXzsCNmhgPaDdNCp5nPF2"},"modeInfo":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gasLimit":"0","payer":"","granter":""},"tip":null},"signatures":["hx8Qo6xZ/Ie0d1TFtiVxSK1rUsRKDEiv1IdcgbkSGYgePYZl6aHJxpSxQDXdIeoZiPeIdrsTkkgjmH4wv2BBdw=="]}`),
			fileFormat: "json",
		},
		{
			name:       "signMode direct text",
			digest:     []byte("body:{messages:{[/offchain.MsgSignArbitraryData]:{appDomain:\"simd\" signer:\"cosmos15r8vphexk8tnu6gvq0a5dhfs3j06ht9kux78rp\" data:\"{\\n\\t\\\"name\\\": \\\"John\\\",\\n\\t\\\"surname\\\": \\\"Connor\\\",\\n\\t\\\"age\\\": 15\\n}\\n\"}}} auth_info:{signer_infos:{public_key:{[/cosmos.crypto.secp256k1.PubKey]:{key:\"\\x03\\xa85\\xb37ړO\\xd6x\\xb5\\x9f\\xb1\\x83\\x92`\\x1b\\xf7Q\\xd0<v\\xdf\\xc4}\\x82\\xcb\\x1eG\\xf5c\\x9c\\xad\"}} mode_info:{single:{mode:SIGN_MODE_DIRECT}}} fee:{}} signatures:\"\\x05\\xacz\\xfa1\\xba\\xa4d\\xc8\\xfa\\xdcT\\xa8B7\\xa4\\xc7餣jf\\xee\\x1e\\xecp\\x07\\xfe\\xb61\\\"Fd\\x19z\\x89(8&\\xf0J\\xe2\\xdd\\\"C\\xe8\\x7ffH5\\r\\xd8E\\xb5TH\\x80v\\x9dNew:\\x03\""),
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
