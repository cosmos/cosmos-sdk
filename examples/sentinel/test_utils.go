package sentinel

import (
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"encoding/hex"
)

var (
	pks1="baymax19@baymax19-HP-Laptop-15-bs1xx:baymax19@baymax19privatekey"
	pks2="sdkldsjgkdmgkgjkrmhnbjrtlweitewotkldmgoirjetkrglmnsjakdhgdskamdg"


	pvk1 = crypto.GenPrivKeyEd25519()
	pvk2 = crypto.GenPrivKeyEd25519()
	pvk3 = crypto.GenPrivKeyEd25519()

	pk1        = pvk1.PubKey()
	pk2        = pvk2.PubKey()
	pk3        = pvk3.PubKey()
	p1, p2, p3 = pk1, pk2, pk3
	addr1      = sdk.AccAddress(p1.Address())
	addr2      = sdk.AccAddress(p2.Address())
	addr3      = sdk.AccAddress(p3.Address())

	emptypk   crypto.PubKey
	emptyaddr sdk.AccAddress
)
var (
	coinPos  = sdk.NewCoin("sentinelToken", 10)
	coinZero = sdk.NewCoin("sentinelToken", 0)
	coinNeg  = sdk.NewCoin("sentinelToken", -1)
)

var (
	cadd1 = sdk.AccAddress("sentinel")
	cadd2 = sdk.AccAddress("cosmos")
	cadd3 = sdk.AccAddress("tendermint")

	kb, err = keys.GetKeyBase()
)
var priv1 crypto.PrivKeyEd25519
var priv2 crypto.PrivKeyEd25519
var priv3 crypto.PrivKeyEd25519

func GetPrivateKey1() crypto.PrivKey{
	pk,_:=hex.DecodeString(pks1)
	copy(priv1[:],pk)
	return priv1
}
func GetPrivateKey2() crypto.PrivKey{
	pk2,_:=hex.DecodeString(pks2)
	copy(priv2[:],pk2)
	return priv2
}
func GetPubkey1()crypto.PubKey{
	pk,_:=hex.DecodeString(pks1)
	copy(priv1[:],pk)
	return priv1.PubKey()
}
func GetAddress1() sdk.AccAddress{
	pk,_:=hex.DecodeString(pks1)
	copy(priv1[:],pk)
	return sdk.AccAddress(priv1.PubKey().Address())

}
func GetPubkey2()crypto.PubKey{
	pk,_:=hex.DecodeString(pks2)
	copy(priv2[:],pk)
	return priv2.PubKey()
}
func GetAddress2() sdk.AccAddress{
	pk,_:=hex.DecodeString(pks2)
	copy(priv2[:],pk)
	return sdk.AccAddress(priv2.PubKey().Address())

}