package ledger

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	cryptoAmino "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestErrorHandling(t *testing.T) {
	// first, try to generate a key, must return an error
	// (no panic)
	path := *hd.NewParams(44, 555, 0, false, 0)
	_, err := NewPrivKeySecp256k1Unsafe(path)
	require.Error(t, err)
}

func TestPublicKeyUnsafe(t *testing.T) {
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, err := NewPrivKeySecp256k1Unsafe(path)
	require.Nil(t, err, "%s", err)
	require.NotNil(t, priv)

	require.Equal(t, "eb5ae9870a21034fef9cd7c4c63588d3b03feb5281b9d232cba34d6f3d71aee59211ffbfe1fe87",
		fmt.Sprintf("%x", cdc.Amino.MustMarshalBinaryBare(priv.PubKey())),
		"Is your device using test mnemonic: %s ?", testutil.TestMnemonic)

	pubKeyAddr, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, priv.PubKey())
	require.NoError(t, err)
	require.Equal(t, "cosmospub1addwnpc2yyp5lmuu6lzvvdvg6wcrl66jsxuayvkt5dxk70t34mjeyy0lhlslapcuuuc4e",
		pubKeyAddr, "Is your device using test mnemonic: %s ?", testutil.TestMnemonic)

	addr := sdk.AccAddress(priv.PubKey().Address()).String()
	require.Equal(t, "cosmos1w34k53py5v5xyluazqpq65agyajavep2rflq6h",
		addr, "Is your device using test mnemonic: %s ?", testutil.TestMnemonic)
}

func TestPublicKeyUnsafeHDPath(t *testing.T) {
	expectedAnswers := []string{
		"cosmospub1addwnpc2yyp5lmuu6lzvvdvg6wcrl66jsxuayvkt5dxk70t34mjeyy0lhlslapcuuuc4e",
		"cosmospub1addwnpc2yypxp5zg0g7le6fz3mhz6rvr5s8kzv0429fxerjjqeh70ls7ffgfveswmjqjd",
		"cosmospub1addwnpc2yyp6yecrj0gzk93dpmgx5zqyr6qds6lrdszkgv6j2n0hgcjy06mf4vce4r7kc",
		"cosmospub1addwnpc2yypnyghuv9u4qamezej4gj5sws8gattr3gu35wu0jfslfg3xkwtvqssersfgs",
		"cosmospub1addwnpc2yypl2a68xdyd0vq70te0y30rdwvdrqdujd0v3d2jehjex2mydhrmupq6s3ks3",
		"cosmospub1addwnpc2yypz9vd9fp47pgk470zcv6lydczar00gekj758zvw75mcjxjlgn480qppwn5p",
		"cosmospub1addwnpc2yyph0gwgymf6q09ya620cn02d0xt9wk972kqgxdp9rpflr5g78ljjks4aax42",
		"cosmospub1addwnpc2yyp3kawgg3fexk4hd7xg6zm9vmplesgpe3w9n4cqp07fzqvkr6fs3kgfmf902",
		"cosmospub1addwnpc2yypcjpdyysemr4nhej906d5xzsctn2zjj9cmqct0wvm9nuf3c0uqyggfn76lm",
		"cosmospub1addwnpc2yypchelnfzgzmrpqhjydxg557nemsxfgg4ypyg3fmmx34hc606cgfzcr2rhhp",
	}

	const numIters = 10

	privKeys := make([]tmcrypto.PrivKey, numIters)

	// Check with device
	for i := uint32(0); i < 10; i++ {
		path := *hd.NewFundraiserParams(0, sdk.CoinType, i)
		fmt.Printf("Checking keys at %v\n", path)

		priv, err := NewPrivKeySecp256k1Unsafe(path)
		require.Nil(t, err, "%s", err)
		require.NotNil(t, priv)

		// Check other methods
		require.NoError(t, priv.(PrivKeyLedgerSecp256k1).ValidateKey())
		tmp := priv.(PrivKeyLedgerSecp256k1)
		(&tmp).AssertIsPrivKeyInner()

		pubKeyAddr, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, priv.PubKey())
		require.NoError(t, err)
		require.Equal(t,
			expectedAnswers[i], pubKeyAddr,
			"Is your device using test mnemonic: %s ?", testutil.TestMnemonic)

		// Store and restore
		serializedPk := priv.Bytes()
		require.NotNil(t, serializedPk)
		require.True(t, len(serializedPk) >= 50)

		privKeys[i] = priv
	}

	// Now check equality
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			require.Equal(t, i == j, privKeys[i].Equals(privKeys[j]))
			require.Equal(t, i == j, privKeys[j].Equals(privKeys[i]))
		}
	}
}

func TestPublicKeySafe(t *testing.T) {
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, addr, err := NewPrivKeySecp256k1(path, "cosmos")

	require.Nil(t, err, "%s", err)
	require.NotNil(t, priv)

	require.Nil(t, ShowAddress(path, priv.PubKey(), sdk.GetConfig().GetBech32AccountAddrPrefix()))

	require.Equal(t, "eb5ae9870a21034fef9cd7c4c63588d3b03feb5281b9d232cba34d6f3d71aee59211ffbfe1fe87",
		fmt.Sprintf("%x", cdc.Amino.MustMarshalBinaryBare(priv.PubKey())),
		"Is your device using test mnemonic: %s ?", testutil.TestMnemonic)

	pubKeyAddr, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, priv.PubKey())
	require.NoError(t, err)
	require.Equal(t, "cosmospub1addwnpc2yyp5lmuu6lzvvdvg6wcrl66jsxuayvkt5dxk70t34mjeyy0lhlslapcuuuc4e",
		pubKeyAddr, "Is your device using test mnemonic: %s ?", testutil.TestMnemonic)

	require.Equal(t, "cosmos1w34k53py5v5xyluazqpq65agyajavep2rflq6h",
		addr, "Is your device using test mnemonic: %s ?", testutil.TestMnemonic)

	addr2 := sdk.AccAddress(priv.PubKey().Address()).String()
	require.Equal(t, addr, addr2)
}

func TestPublicKeyHDPath(t *testing.T) {
	expectedPubKeys := []string{
		"cosmospub1addwnpc2yyp5lmuu6lzvvdvg6wcrl66jsxuayvkt5dxk70t34mjeyy0lhlslapcuuuc4e",
		"cosmospub1addwnpc2yypxp5zg0g7le6fz3mhz6rvr5s8kzv0429fxerjjqeh70ls7ffgfveswmjqjd",
		"cosmospub1addwnpc2yyp6yecrj0gzk93dpmgx5zqyr6qds6lrdszkgv6j2n0hgcjy06mf4vce4r7kc",
		"cosmospub1addwnpc2yypnyghuv9u4qamezej4gj5sws8gattr3gu35wu0jfslfg3xkwtvqssersfgs",
		"cosmospub1addwnpc2yypl2a68xdyd0vq70te0y30rdwvdrqdujd0v3d2jehjex2mydhrmupq6s3ks3",
		"cosmospub1addwnpc2yypz9vd9fp47pgk470zcv6lydczar00gekj758zvw75mcjxjlgn480qppwn5p",
		"cosmospub1addwnpc2yyph0gwgymf6q09ya620cn02d0xt9wk972kqgxdp9rpflr5g78ljjks4aax42",
		"cosmospub1addwnpc2yyp3kawgg3fexk4hd7xg6zm9vmplesgpe3w9n4cqp07fzqvkr6fs3kgfmf902",
		"cosmospub1addwnpc2yypcjpdyysemr4nhej906d5xzsctn2zjj9cmqct0wvm9nuf3c0uqyggfn76lm",
		"cosmospub1addwnpc2yypchelnfzgzmrpqhjydxg557nemsxfgg4ypyg3fmmx34hc606cgfzcr2rhhp",
	}

	expectedAddrs := []string{
		"cosmos1w34k53py5v5xyluazqpq65agyajavep2rflq6h",
		"cosmos19ewxwemt6uahejvwf44u7dh6tq859tkyvarh2q",
		"cosmos1a07dzdjgjsntxpp75zg7cgatgq0udh3pcdcxm3",
		"cosmos1qvw52lmn9gpvem8welghrkc52m3zczyhlqjsl7",
		"cosmos17m78ka80fqkkw2c4ww0v4xm5nsu2drgrlm8mn2",
		"cosmos1ferh9ll9c452d2p8k2v7heq084guygkn43up9e",
		"cosmos10vf3sxmjg96rqq36axcphzfsl74dsntuehjlw5",
		"cosmos1cq83av8cmnar79h0rg7duh9gnr7wkh228a7fxg",
		"cosmos1dszhfrt226jy5rsre7e48vw9tgwe90uerfyefa",
		"cosmos1734d7qsylzrdt05muhqqtpd90j8mp4y6rzch8l",
	}

	const numIters = 10

	privKeys := make([]tmcrypto.PrivKey, numIters)

	// Check with device
	for i := uint32(0); i < 10; i++ {
		path := *hd.NewFundraiserParams(0, sdk.CoinType, i)
		fmt.Printf("Checking keys at %v\n", path)

		priv, addr, err := NewPrivKeySecp256k1(path, "cosmos")
		require.Nil(t, err, "%s", err)
		require.NotNil(t, addr)
		require.NotNil(t, priv)

		addr2 := sdk.AccAddress(priv.PubKey().Address()).String()
		require.Equal(t, addr2, addr)
		require.Equal(t,
			expectedAddrs[i], addr,
			"Is your device using test mnemonic: %s ?", testutil.TestMnemonic)

		// Check other methods
		require.NoError(t, priv.(PrivKeyLedgerSecp256k1).ValidateKey())
		tmp := priv.(PrivKeyLedgerSecp256k1)
		(&tmp).AssertIsPrivKeyInner()

		pubKeyAddr, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, priv.PubKey())
		require.NoError(t, err)
		require.Equal(t,
			expectedPubKeys[i], pubKeyAddr,
			"Is your device using test mnemonic: %s ?", testutil.TestMnemonic)

		// Store and restore
		serializedPk := priv.Bytes()
		require.NotNil(t, serializedPk)
		require.True(t, len(serializedPk) >= 50)

		privKeys[i] = priv
	}

	// Now check equality
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			require.Equal(t, i == j, privKeys[i].Equals(privKeys[j]))
			require.Equal(t, i == j, privKeys[j].Equals(privKeys[i]))
		}
	}
}

func getFakeTx(accountNumber uint32) []byte {
	tmp := fmt.Sprintf(
		`{"account_number":"%d","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"5000"},"memo":"memo","msgs":[[""]],"sequence":"6"}`,
		accountNumber)

	return []byte(tmp)
}

func TestSignaturesHD(t *testing.T) {
	for account := uint32(0); account < 100; account += 30 {
		msg := getFakeTx(account)

		path := *hd.NewFundraiserParams(account, sdk.CoinType, account/5)
		fmt.Printf("Checking signature at %v    ---   PLEASE REVIEW AND ACCEPT IN THE DEVICE\n", path)

		priv, err := NewPrivKeySecp256k1Unsafe(path)
		require.Nil(t, err, "%s", err)

		pub := priv.PubKey()
		sig, err := priv.Sign(msg)
		require.Nil(t, err)

		valid := pub.VerifySignature(msg, sig)
		require.True(t, valid, "Is your device using test mnemonic: %s ?", testutil.TestMnemonic)
	}
}

func TestRealDeviceSecp256k1(t *testing.T) {
	msg := getFakeTx(50)
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, err := NewPrivKeySecp256k1Unsafe(path)
	require.Nil(t, err, "%s", err)

	pub := priv.PubKey()
	sig, err := priv.Sign(msg)
	require.Nil(t, err)

	valid := pub.VerifySignature(msg, sig)
	require.True(t, valid)

	// now, let's serialize the public key and make sure it still works
	bs := cdc.Amino.MustMarshalBinaryBare(priv.PubKey())
	pub2, err := cryptoAmino.PubKeyFromBytes(bs)
	require.Nil(t, err, "%+v", err)

	// make sure we get the same pubkey when we load from disk
	require.Equal(t, pub, pub2)

	// signing with the loaded key should match the original pubkey
	sig, err = priv.Sign(msg)
	require.Nil(t, err)
	valid = pub.VerifySignature(msg, sig)
	require.True(t, valid)

	// make sure pubkeys serialize properly as well
	bs = cdc.Amino.MustMarshalBinaryBare(pub)
	bpub, err := cryptoAmino.PubKeyFromBytes(bs)
	require.NoError(t, err)
	require.Equal(t, pub, bpub)
}
