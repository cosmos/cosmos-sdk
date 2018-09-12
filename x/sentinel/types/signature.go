package types

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/bech32"
	"fmt"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
)

type ClientSignature struct {
	Coins     sdk.Coins
	Sessionid []byte
	Counter   int64
	Signature Signature
	IsFinal   bool
}
type Signature struct {
	Pubkey    crypto.PubKey    `json:"pub_key"`
	Signature crypto.Signature `json:"signature"`
}

func NewClientSignature(coins sdk.Coins, sesid []byte, counter int64, pubkey crypto.PubKey, sign crypto.Signature, isfinal bool) ClientSignature {
	return ClientSignature{
		Coins:     coins,
		Sessionid: sesid,
		Counter:   counter,
		IsFinal:   isfinal,
		Signature: Signature{
			Pubkey:    pubkey,
			Signature: sign,
		},
	}
}
func (a ClientSignature) Value() Signature {
	return a.Signature
}

type StdSig struct {
	Coins     sdk.Coins
	Sessionid []byte
	Counter   int64
	Isfinal   bool
}

func ClientStdSignBytes(coins sdk.Coins, sessionid []byte, counter int64, isfinal bool) []byte {
	bz, err := json.Marshal(StdSig{
		Coins:     coins,
		Sessionid: sessionid,
		Counter:   counter,
		Isfinal:   isfinal,
	})
	if err != nil {
	}
	return sdk.MustSortJSON(bz)
}

type Vpnsign struct {
	From     sdk.AccAddress
	Ip       string
	Netspeed int64
	Ppgb     int64
	Location string
}

func GetVPNSignature(address sdk.AccAddress, ip string, ppgb int64, netspeed int64, location string) []byte {
	bz, err := json.Marshal(Vpnsign{
		From:     address,
		Ip:       ip,
		Ppgb:     ppgb,
		Netspeed: netspeed,
		Location: location,
	})
	if err != nil {

	}
	return sdk.MustSortJSON(bz)
}
func GetBech32Signature(sign crypto.Signature) (string, error) {
	return bech32.ConvertAndEncode("", sign.Bytes())

}

func GetBech64Signature(address string) (pk crypto.Signature, err error) {
	hrp, bz, err := DecodeAndConvert(address)
	if err != nil {
		return nil, err
	}
	if hrp != "" {
		return nil, fmt.Errorf("invalid bech32 prefix. Expected %s, Got %s", "", hrp)
	}

	pk, err = cryptoAmino.SignatureFromBytes(bz)

	if err != nil {
		return nil, err
	}

	return pk, nil
}

