package sentinel

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	ms, authkey, sentkey = CreateMultiStore()
	cdc                  = wire.NewCodec()
	ac                   = auth.NewAccountMapper(cdc, authkey, auth.ProtoBaseAccount)
	ctx                  = sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	keeper               = NewKeeper(cdc, sentkey, bank.NewKeeper(ac), ac, DefaultCodeSpace)
	cd, k, c, am         = cdc, keeper, ctx, ac
)

func CreateMultiStore() (sdk.MultiStore, *sdk.KVStoreKey, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	authkey := sdk.NewKVStoreKey("authkey")
	sentinelkey := sdk.NewKVStoreKey("sentinel")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authkey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(sentinelkey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, authkey, sentinelkey

}

func TestRegisterVpnService(t *testing.T) {

	msg, _ := k.RegisterVpnService(c, MsgRegisterVpnService{From: GetAddress1(), Ip: "127.0.0.1", Netspeed: 12, Ppgb: 19, Location: "HYd"})
	t.Log(len(GetPubkey1().Bytes()))
	require.Equal(t, msg, GetAddress2())
}

func TestPayVpnService(t *testing.T) {

	auth.RegisterBaseAccount(cd)
	aut1 := am.NewAccountWithAddress(c, GetAddress1())
	aut1.SetPubKey(GetPubkey1())
	aut1.SetCoins(sdk.Coins{coinPos})
	am.SetAccount(c, aut1)

	aut2 := am.NewAccountWithAddress(c, GetAddress2())
	aut2.SetCoins(sdk.Coins{coinPos})
	aut2.SetPubKey(GetPubkey2())
	am.SetAccount(c, aut2)

	msg, _ := k.RegisterVpnService(c, MsgRegisterVpnService{From: GetAddress1(), Ip: "127.0.0.1", Netspeed: 12, Ppgb: 19, Location: "HYd"})
	require.Equal(t, msg, GetAddress1())

	b, _, _ := k.PayVpnService(c, MsgPayVpnService{sdk.Coins{coinPos}, GetAddress1(), GetAddress2()})
	require.Equal(t, len(b), 20)
	require.NotNil(t, b)
	t.Log(cd, k, c, am)
}

/*
func TestGetVpnPayment(t *testing.T) {


	auth.RegisterBaseAccount(cd)
	aut1 := am.NewAccountWithAddress(c, GetAddress1())
	aut1.SetPubKey(GetPubkey1())
	aut1.SetCoins(sdk.Coins{coinPos})
	am.SetAccount(c, aut1)

	aut2 := am.NewAccountWithAddress(c, GetAddress2())
	aut2.SetCoins(sdk.Coins{coinPos})
	aut2.SetPubKey(GetPubkey2())
	am.SetAccount(c, aut2)

	msg, _ := k.RegisterVpnService(c, MsgRegisterVpnService{From: GetAddress1(), Ip: "127.0.0.1", Netspeed: 12, Ppgb: 19, Location: "HYd",Signature:sign1,Pubkey:pk1})
	require.Equal(t, msg, GetAddress1())

	b, _ := k.PayVpnService(c, MsgPayVpnService{sdk.Coins{coinPos},GetAddress1(),GetAddress2()})
	require.Equal(t, len(b),20)
	require.NotNil(t, b)
	t.Log(cd, k, c, am)

	sessionid := []byte(b)

	require.Nil(t, err)

	clientsession := senttype.GetNewSessionMap(coinPos, pk2, pk1)
	bz := senttype.ClientStdSignBytes(coinPos, sessionid, 1, false)
	t.Log(bz)
	sign1, err = pvk1.Sign(bz)
	t.Log(ctx, keeper)
	t.Log(sign1)
	mg := MsgGetVpnPayment{
		Signature: sign1,
		Coins:     coinPos,
		Sessionid: sessionid,
		Counter:   1,
		Pubkey:    pk1,
		From:      addr2,
		IsFinal:   false,
	}
	a, err := keeper.GetVpnPayment(ctx, MsgGetVpnPayment{Signature: sign1, Coins: coinPos, Sessionid: sessionid, Pubkey: pk1, From: addr2, IsFinal: false, Counter: 1})
	require.Nil(t, err)
	require.Equal(t, sessionid, a)
}
*/
