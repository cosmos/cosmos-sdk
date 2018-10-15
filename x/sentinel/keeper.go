package sentinel

import (
	"crypto/md5"
	"encoding/hex"
	"math"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	senttype "github.com/cosmos/cosmos-sdk/x/sentinel/types"
	"github.com/tendermint/tendermint/crypto"
)

type PubKeyEd25519 [32]byte

type Keeper struct {
	sentStoreKey sdk.StoreKey
	coinKeeper   bank.Keeper
	cdc          *wire.Codec

	codespace sdk.CodespaceType
	account   auth.AccountMapper
}

type Sign struct {
	coin    sdk.Coin
	vpnaddr sdk.AccAddress
	counter int64
	hash    string
	sign    crypto.PubKey
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper, am auth.AccountMapper, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		sentStoreKey: key,
		cdc:          cdc,
		coinKeeper:   ck,
		codespace:    codespace,
		account:      am,
	}

}

func (keeper Keeper) RegisterVpnService(ctx sdk.Context, msg MsgRegisterVpnService) (sdk.AccAddress, sdk.Error) {
	sentKey := msg.From
	store := ctx.KVStore(keeper.sentStoreKey)
	address := store.Get([]byte(sentKey))
	if address == nil {
		vpnreg := senttype.NewVpnRegister(msg.Ip, msg.NetSpeed.UploadSpeed, msg.NetSpeed.DownloadSpeed, msg.PricePerGb, msg.EncMethod, msg.Location.Latitude, msg.Location.Longitude, msg.Location.City, msg.Location.Country, msg.NodeType, msg.Version)
		bz, _ := keeper.cdc.MarshalBinary(vpnreg)
		store.Set([]byte(sentKey), bz)
		return msg.From, nil
	}
	return nil, ErrAccountAddressExist("Address already RegsentKeyistered as VPN node")
}

func (keeper Keeper) RegisterMasterNode(ctx sdk.Context, msg MsgRegisterMasterNode) (sdk.AccAddress, sdk.Error) {
	sentkey := msg.Address
	store := ctx.KVStore(keeper.sentStoreKey)
	address := store.Get([]byte(sentkey))
	if address == nil {
		address := msg.Address
		bz, _ := keeper.cdc.MarshalBinary(address)
		store.Set([]byte(msg.Address), bz)
		return msg.Address, nil
	}
	return nil, ErrAccountAddressExist("Address already registered as MasterNode")
}

func (keeper Keeper) StoreKey() sdk.StoreKey {
	return keeper.sentStoreKey
}

func (keeper Keeper) DeleteVpnService(ctx sdk.Context, msg MsgDeleteVpnUser) (sdk.AccAddress, sdk.Error) {

	store := ctx.KVStore(keeper.sentStoreKey)
	db := store.Get([]byte(msg.Vaddr))
	if db == nil {
		return nil, ErrAccountAddressNotExist("Account is not exist")
	}
	store.Delete([]byte(msg.Vaddr))
	return msg.Vaddr, nil
}
func (keeper Keeper) DeleteMasterNode(ctx sdk.Context, msg MsgDeleteMasterNode) (sdk.AccAddress, sdk.Error) {
	store := ctx.KVStore(keeper.sentStoreKey)
	db := store.Get([]byte(msg.Maddr))
	if db == nil {
		return nil, ErrAccountAddressNotExist("Account is not exist")
	}
	store.Delete([]byte(msg.Maddr))
	return msg.Maddr, nil
}

func (keeper Keeper) PayVpnService(ctx sdk.Context, msg MsgPayVpnService) (string, sdk.Error) {

	var err error
	hash := md5.New()
	sequence, err := keeper.account.GetSequence(ctx, msg.From)
	if err != nil {
		return "", sdk.ErrInvalidSequence("Invalid sequence")
	}
	addressbytes := []byte(msg.From.String() + "" + strconv.Itoa(int(sequence)))
	hash.Write(addressbytes)
	if err != nil {
		return "", ErrBech32Decode("address hash is failed")
	}
	sentKey := hex.EncodeToString(hash.Sum(nil))[:20]
	vpnpub, err := keeper.account.GetPubKey(ctx, msg.Vpnaddr)
	if err != nil {
		return "", ErrInvalidPubKey("Vpn pubkey failed")
	}
	time := ctx.BlockHeader().Time
	session := senttype.GetNewSessionMap(msg.Coins, vpnpub, msg.Pubkey, msg.From, time)
	store := ctx.KVStore(keeper.sentStoreKey)
	data := store.Get([]byte(msg.Vpnaddr))
	if data == nil {
		return "", sdk.ErrUnknownAddress("VPN address is not registered")
	}
	bz, err := keeper.cdc.MarshalBinary(session)
	if err != nil {
		return "", ErrMarshal("Marshal of session struct is failed")
	}
	_, _, err = keeper.coinKeeper.SubtractCoins(ctx, msg.From, msg.Coins)
	if err != nil {
		return "", sdk.ErrInsufficientCoins("Coins Parse failed or insufficient funds")
	}
	store.Set([]byte(sentKey), bz)
	return string(sentKey[:]), nil
}
func (keeper Keeper) RefundBal(ctx sdk.Context, msg MsgRefund) (sdk.AccAddress, sdk.Error) {

	var err error
	var clientSession senttype.Session
	store := ctx.KVStore(keeper.sentStoreKey)
	x := store.Get([]byte(msg.Sessionid))
	if x == nil {
		return nil, ErrInvalidSessionid("Invalid SessionId")
	}
	err = keeper.cdc.UnmarshalBinary(x, &clientSession)
	if err != nil {
		return nil, ErrUnMarshal("UnMarshal of bytes failed")

	}
	caddr := sdk.AccAddress(clientSession.CAddress)
	if msg.From.String() != caddr.String() {
		return nil, sdk.ErrUnknownAddress("Address is not associated with this Session")
	}
	ctime := ctx.BlockHeader().Time
	if clientSession.Status == 0 {
		_, _, err = keeper.coinKeeper.AddCoins(ctx, msg.From, clientSession.TotalLockedCoins.Minus(clientSession.ReleasedCoins))
		if err != nil {
			return nil, sdk.ErrInsufficientCoins("Insufficient funds")
		}
		store.Delete(msg.Sessionid)
		return msg.From, nil
	}
	if clientSession.Status == 1 {
		time := int64(math.Abs(float64(ctime))) - clientSession.Timestamp
		if time >= 86400 && clientSession.TotalLockedCoins.Minus(clientSession.ReleasedCoins).IsPositive() && !clientSession.TotalLockedCoins.Minus(clientSession.ReleasedCoins).IsZero() {
			_, _, err = keeper.coinKeeper.AddCoins(ctx, msg.From, clientSession.TotalLockedCoins.Minus(clientSession.ReleasedCoins))
			if err != nil {
				return nil, sdk.ErrInsufficientCoins("Insufficient funds")
			}
			store.Delete(msg.Sessionid)
			return msg.From, nil
		}
		return nil, ErrTimeInterval("time is less than 24 hours  or the balance is negative or equal to zero")
	}
	return nil, ErrInvalidSessionid("Invalid SessionId")

}
func (keeper Keeper) GetVpnPayment(ctx sdk.Context, msg MsgGetVpnPayment) ([]byte, sdk.AccAddress, sdk.Error) {

	var clientSession senttype.Session

	store := ctx.KVStore(keeper.sentStoreKey)
	x := store.Get([]byte(msg.Sessionid))
	if x == nil {
		return nil, nil, ErrInvalidSessionid("Invalid session Id")
	}
	err := keeper.cdc.UnmarshalBinary(x, &clientSession)
	if err != nil {
		return nil, nil, ErrUnMarshal("Unmarshal of bytes failed")
	}
	ClientPubkey := clientSession.CPubKey
	signBytes := senttype.ClientStdSignBytes(msg.Coins, []byte(msg.Sessionid), msg.Counter, msg.IsFinal)
	if !ClientPubkey.VerifyBytes(signBytes, msg.Signature) {
		return nil, nil, sdk.ErrUnauthorized("signature verification failed")
	}
	clientSessionData := clientSession
	if msg.Counter > clientSessionData.Counter {
		clientSessionData.Counter = msg.Counter
		CoinsToAdd := msg.Coins.Minus(clientSessionData.ReleasedCoins)
		if !CoinsToAdd.IsZero() && (clientSessionData.TotalLockedCoins.Minus(clientSessionData.ReleasedCoins)).Minus(CoinsToAdd).IsPositive() && !CoinsToAdd.IsZero() {
			clientSessionData.ReleasedCoins = msg.Coins
			VpnAddr := sdk.AccAddress(clientSessionData.VpnPubKey.Address())
			_, _, err = keeper.coinKeeper.AddCoins(ctx, VpnAddr, CoinsToAdd)
			if err != nil {
				return nil, nil, sdk.ErrInsufficientCoins("Insufficient funds")
			}
			sentKey := []byte(msg.Sessionid)

			if clientSessionData.TotalLockedCoins.Minus(clientSessionData.ReleasedCoins).IsZero() && !clientSessionData.TotalLockedCoins.Minus(clientSessionData.ReleasedCoins).IsPositive() || clientSessionData.Status == 0 {
				store.Delete(sentKey)
				return nil, sentKey, sdk.ErrInsufficientCoins("Insufficient funds")
			}

			if msg.IsFinal == true {
				clientSessionData.Status = 0
				bz, err := keeper.cdc.MarshalBinary(clientSessionData)
				if err != nil {
					return nil, nil, ErrUnMarshal("Unmarshal of bytes failed")
				}
				store.Set(sentKey, bz)
				clientAddr, err := keeper.RefundBal(ctx, MsgRefund{From: clientSessionData.CAddress, Sessionid: msg.Sessionid})
				if err != nil {
					return nil, nil, sdk.ErrInternal("Refund failed")
				}
				return clientAddr, sentKey, nil
			}
			bz, err := keeper.cdc.MarshalBinary(clientSessionData)
			if err != nil {
				return nil, nil, ErrUnMarshal("Unmarshal of bytes failed")
			}
			store.Set(sentKey, bz)

			return nil, sentKey, nil
		}
		return nil, msg.Sessionid, sdk.ErrInsufficientCoins("Insufficient Coins ")
	}
	return nil, msg.Sessionid, ErrSignMsg("Invalid Counter")
}
func (keeper Keeper) NewMsgDecoder(acc []byte) (senttype.Registervpn, sdk.Error) {

	msg := senttype.Registervpn{}
	err := keeper.cdc.UnmarshalBinary(acc, &msg)
	if err != nil {
		return msg, ErrUnMarshal("unmarshal failed")
	}
	return msg, ErrUnMarshal("Unmarshal of bytes failed")

}

func (keeper Keeper) SendTokens(ctx sdk.Context, msg MsgSendTokens) (sdk.AccAddress, sdk.Error) {
	_, _, err := keeper.coinKeeper.SubtractCoins(ctx, msg.From, msg.Coins)
	if err != nil {
		return nil, sdk.ErrInsufficientFunds("Insufficient funds")
	}
	_, _, err = keeper.coinKeeper.AddCoins(ctx, msg.To, msg.Coins)
	if err != nil {
		return nil, sdk.ErrInsufficientFunds("Failed to retrive funds")
	}
	return msg.To, nil
}

//func (keeper Keeper) GetsentStore(ctx sdk.Context, msg MsgRegisterMasterNode) (sdk.KVStore, sdk.AccAddress) {
//	return ctx.KVStore(keeper.sentStoreKey), addr1
//}
