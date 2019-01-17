package main

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	txBuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bip39 "github.com/cosmos/go-bip39"
	amino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/bech32"
)

func main() {
	var cdc = codec.New()
	bank.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)

	amino.RegisterAmino(cdc)
	crypto.RegisterAmino(cdc)

	rawJsonTx := []byte(`{"type":"auth/StdTx","value":{"msg":[{"type":"cosmos-sdk/Send","value":{"inputs":[{"address":"cosmos13kujs7dzumc0k2vy37s9zs6j5da6qmn6udddza","coins":[{"denom":"STAKE","amount":"10"}]}],"outputs":[{"address":"cosmos1gfzexy3z0qfc97mjudjy5zeg2fje6k442phy6r","coins":[{"denom":"STAKE","amount":"10"}]}]}}],"fee":{"amount":[{"denom":"","amount":"0"}],"gas":"200000"},"signatures":null,"memo":""}}`)
	fmt.Println("[1] : ", rawJsonTx)           // [391/416]0xc42098b6c0
	fmt.Println("[1.1] : ", string(rawJsonTx)) // {"type":"auth/StdTx","value":{"msg":[{"type":"cosmos-sdk/Send","value":{"inputs":[{"address":"cosmos13kujs7dzumc0k2vy37s9zs6j5da6qmn6udddza","coins":[{"denom":"STAKE","amount":"10"}]}],"outputs":[{"address":"cosmos1gfzexy3z0qfc97mjudjy5zeg2fje6k442phy6r","coins":[{"denom":"STAKE","amount":"10"}]}]}}],"fee":{"amount":[{"denom":"","amount":"0"}],"gas":"200000"},"signatures":null,"memo":""}}

	// Wrap rawJsonTx with authStdTx
	var authStdTx auth.StdTx
	cdc.UnmarshalJSON(rawJsonTx, &authStdTx)
	fmt.Println("[2] ", authStdTx, reflect.TypeOf(authStdTx)) // {[{[{8DB92879A2E6F0FB29848FA0514352A37BA06E7A 10STAKE}] [{4245931222781382FB72E3644A0B2852659D5AB5 10STAKE}]}] {0 200000} [] }
	fmt.Println("[2.1] ", reflect.TypeOf(authStdTx))          // auth.StdTx

	// Transaction Builder
	txBldr := txBuilder.NewTxBuilderFromCLI().
		WithAccountNumber(0).
		WithSequence(9).
		WithChainID("game_of_stakes_3")

	// {<nil> 0 9 200000 0 false game_of_stakes_3  []}
	fmt.Println("[3] ", txBldr)

	mnemonic := "iron breeze tongue voice stomach nut manage advance rather mad hurry neutral identify armed unusual crunch hammer scan riot mom surface horn stamp thank"
	seed := bip39.NewSeed(mnemonic, "")
	masterKey, ch := hd.ComputeMastersFromSeed(seed)
	derivedKey, _ := hd.DerivePrivateKeyForPath(masterKey, ch, hd.FullFundraiserPath)
	cosmosAddr, _ := bech32.ConvertAndEncode("cosmos", secp256k1.PrivKeySecp256k1(derivedKey).PubKey().Address())

	// 4245931222781382FB72E3644A0B2852659D5AB5
	fmt.Println("[4] ", secp256k1.PrivKeySecp256k1(derivedKey).PubKey().Address())
	// cosmos1gfzexy3z0qfc97mjudjy5zeg2fje6k442phy6r
	fmt.Println("[5] ", cosmosAddr)

	stdSignMsg := txBuilder.StdSignMsg{
		txBldr.GetChainID(),
		txBldr.GetAccountNumber(),
		txBldr.GetSequence(),
		authStdTx.Fee,
		authStdTx.Msgs,
		authStdTx.Memo,
	}
	// {game_of_stakes_3 0 9 {0 200000} [{[{8DB92879A2E6F0FB29848FA0514352A37BA06E7A 10STAKE}] [{4245931222781382FB72E3644A0B2852659D5AB5 10STAKE}]}] } context.StdSignMsg
	fmt.Println("[6] ", stdSignMsg, reflect.TypeOf(stdSignMsg))

	stdSignMsgBytes, _ := cdc.MarshalJSON(stdSignMsg)

	// [123 34 99 104 97 105 110 95 105 100 34 58 34 103 97 109 101 9.....]
	fmt.Println("[7]", stdSignMsgBytes, reflect.TypeOf(stdSignMsgBytes))

	stdSignatureBytes, err := secp256k1.PrivKeySecp256k1(derivedKey).Sign(stdSignMsg.Bytes())
	if err != nil {
		fmt.Println("err ", err)
	}
	// [96 148 123 170 105 199 65 159 61 156 217 83 93 7 212 194 185 154 177 137 186 181 146 33 153 152 167 7 127 174 10 53 12 84 22 113 156 56 213 245 126 97 112 73 54 205 28 18 202 200 254 91 118 255 7 136 140 29 195 28 112 141 8 98]
	fmt.Println("[8] ", stdSignatureBytes)

	stdSignature := auth.StdSignature{
		secp256k1.PrivKeySecp256k1(derivedKey).PubKey(),
		stdSignatureBytes,
	}
	// {PubKeySecp256k1{022795CA53CC6683EB8DC7718F59ED5A1B75CC1EC97E7D9E5352532CBC7D8AC3A6} [96 148 123 170 10
	fmt.Println("[9] ", stdSignature)

	sigs := recover.GetSignatures()
	sigs = []auth.StdSignature{stdSignature}
	signedStdTx := auth.NewStdTx(recover.GetMsgs(), recover.Fee, sigs, recover.GetMemo())

	//  {[{[{8DB92879A2E6F0FB29848FA0514352A37BA06E7A 10STAKE}] [{4245931222781382FB72E3644A0B2852659D5AB5 10STAKE}]}] {0 200000} [{PubKeySe ...
	fmt.Println("[10] ", signedStdTx)

	amino2, _ := cdc.MarshalJSON(signedStdTx)

	// 7b2274797065223a22617574682f5374645478222c2276616c7565223a7b226d7367223a5b7b2274797065223a22636f736d6f732d73646b2f53656e64222c2276616c7565223a7b22696e70757473223a5b7b2261646472657373223a22636f736d6f7331336b756a7337
	fmt.Println("[11]  : ", amino2)

	fmt.Println("[11.1] : ", hex.EncodeToString(amino2))
	//fmt.Println("amino2 marshaledString : ", string(amino2))

	amino3, _ := cdc.MarshalJSON(stdSignature.PubKey)
	fmt.Println("[11.2]  : ", amino3)
	fmt.Println("[11.3] : ", hex.EncodeToString(amino3))
}
