package tx

/*
func TestBuilders(t *testing.T) {
	const keyHex = "8c7e006440ac5e358739bdc3d10a8b2d229e23d27660f6d3a8306cee4379594c"
	pkBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		t.Fatal(err)
	}
	pk := secp256k1.PrivKey{Key: pkBytes}
	tmRPC, err := tmrpc.New("tcp://localhost:26657", "")
	if err != nil {
		t.Fatal(err)
	}

	builder := NewUnsignedTxBuilder()
	builder.SetMemo("")
	builder.SetChainID("testing")
	builder.AddMsg(&bank.MsgSend{
		FromAddress: "cosmos1ujtnemf6jmfm995j000qdry064n5lq854gfe3j",
		ToAddress:   "cosmos1caa3es6q3mv8t4gksn9wjcwyzw7cnf5gn5cx7j",
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 10)),
	})
	builder.SetFeePayer("cosmos1ujtnemf6jmfm995j000qdry064n5lq854gfe3j")
	builder.SetFees(sdk.NewCoins(sdk.NewInt64Coin("stake", 10)))
	builder.SetGasLimit(2500000)
	builder.AddSigner(SignerInfo{
		PubKey:        pk.PubKey(),
		SignMode:      signing.SignMode_SIGN_MODE_DIRECT,
		AccountNumber: 0,
		Sequence:      2,
	})

	signedbldr, err := builder.SignedBuilder()
	if err != nil {
		t.Fatal(err)
	}

	expectedSig, err := signedbldr.BytesToSign(pk.PubKey())
	if err != nil {
		t.Fatal(err)
	}
	signedSig, err := pk.Sign(expectedSig)
	if err != nil {
		t.Fatal(err)
	}
	err = signedbldr.SetSignature(pk.PubKey(), signedSig)
	if err != nil {
		t.Fatal(err)
	}

	txB, err := signedbldr.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := tmRPC.BroadcastTxCommit(context.TODO(),txB )
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", resp)
}

*/
