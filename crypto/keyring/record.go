package keyring

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
)

var ErrPrivKeyExtr = errors.New("Private key extraction works only for Local")
//TODO replace Info by reyring entry in client/reys
// check  NewLedgerInfo, newLocalInfo, newMultiInfo in whole codebase
// TODO count how many times NewLedgerInfo or newLocalInfo is used and perhaps consider create a separate functions for that
func NewRecord(name string, pk cryptotypes.PubKey, item isRecord_Item) (*Record, error) {
	any, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return nil, err
	}
	return &Record{name, any, item}, nil
}

// bz, err := cdc.Marshal(privKey) yields an error that's why I cast privKey to curve PrivKey to serialize it
func NewLocalRecord(cdc codec.Codec, privKey cryptotypes.PrivKey) (*Record_Local, error) {

	var (
		err error
		bz  []byte
	)

	privKeyType := privKey.Type()

	// TODO find out why i canot do that?
	b, err := cdc.Marshal(privKey)

	switch privKeyType {
	case "secp256k1":
		priv, ok := privKey.(*secp256k1.PrivKey)
		if !ok {
			return nil, fmt.Errorf("unable to cast privKey to *secp256k1.PrivKey")
		}

		bz, err = cdc.Marshal(priv)
		if err != nil {
			return nil, err
		}

	case "ed25519":
		priv, ok := privKey.(*ed25519.PrivKey)
		if !ok {
			return nil, fmt.Errorf("unable to cast privKey to *ed25519.PrivKey")
		}

		bz, err = cdc.Marshal(priv)
		if err != nil {
			return nil, err
		}
	}

	return &Record_Local{string(bz), privKeyType}, nil
}

func NewLocalRecordItem(localRecord *Record_Local) *Record_Local_ {
	return &Record_Local_{localRecord}
}

func NewLedgerRecord(path *hd.BIP44Params) *Record_Ledger {
	return &Record_Ledger{path}
}

func NewLedgerRecordItem(ledgerRecord *Record_Ledger) *Record_Ledger_ {
	return &Record_Ledger_{ledgerRecord}
}

func (rl *Record_Ledger) GetPath() *hd.BIP44Params {
	return rl.Path
}

func NewEmptyRecord() *Record_Empty {
	return &Record_Empty{}
}

func NewEmptyRecordItem(re *Record_Empty) *Record_Empty_ {
	return &Record_Empty_{re}
}

func (k Record) GetPubKey() (cryptotypes.PubKey, error) {
	pk, ok := k.PubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		// TODO - don't use fmt.Errorf
		return nil, fmt.Errorf("Unable to cast PubKey to cryptotypes.PubKey")
	}
	return pk, nil
}

// GetType implements Info interface
func (k Record) GetAddress() (types.AccAddress, error) {
	pk, err := k.GetPubKey()
	if err != nil {
		return nil, err
	}
	return pk.Address().Bytes(), nil
}

func (k Record) GetAlgo() string {

	if l := k.GetLocal(); l != nil {
		return l.PrivKeyType
	}

	// TODO  doublecheck there is no field pubKeyType for multi,offline,ledger
	return ""
}

// TODO remove it later
func (k Record) GetType() KeyType {
	return 0
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (k *Record) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(k.PubKey, &pk)
}

// encoding info
// we remove tis function aso we can pass cdc.Marrshal install ,we put cdc on reystore
/*
func protoMarshalInfo(i Info) ([]byte, error) {
	re, ok := i.(*Record)
	if !ok {
		return nil, fmt.Errorf("Unable to cast Info to *Record")
	}

	bz, err := proto.Marshal(re)
	if err != nil {
		return nil, sdrerrors.Wrap(err, "Unable to marshal Record to bytes")
	}

	return bz, nil
}
*/

// decoding info
// we remove tis function aso we can pass cdc.Marrshal install ,we put cdc on reystore
/*
func protoUnmarshalInfo(bz []byte, cdc codec.Codec) (Info, error) {

	var k Record // will not work cause we use any, use InterfaceRegistry
	// dont forget to merge master to my branch, UnmarshalBinaryBare has been renamed
	// cdcc.Marshaler.UnmarshalBinaryBare()  // lire proto.UnMarshal but works with Any
	if err := cdc.UnmarshalInterface(bz, &re); err != nil {
		return nil, sdrerrors.Wrap(err, "failed to unmarshal bytes to Info")
	}

	return re, nil
}



func NewBIP44Params(purpose uint32, coinType uint32, account uint32, change bool, adressIndex uint32) *BIP44Params {
	return &BIP44Params{purpose, coinType, account, change, adressIndex}
}

// DerivationPath returns the BIP44 fields as an array.
func (p hd.BIP44Params) DerivationPath() []uint32 {
	change := uint32(0)
	if p.Change {
		change = 1
	}

	return []uint32{
		p.Purpose,
		p.Cointype,
		p.Account,
		change,
		p.Adressindex,
	}
}

func (p BIP44Params) String() string {
	var changeStr string
	if p.Change {
		changeStr = "1"
	} else {
		changeStr = "0"
	}
	return fmt.Sprintf("m/%d'/%d'/%d'/%s/%d",
		p.Purpose,
		p.Cointype,
		p.Account,
		changeStr,
		p.Adressindex)
}
*/

func ExtractPrivKeyFromItem(cdc codec.Codec, k *Record) (cryptotypes.PrivKey, error) {
	rl := k.GetLocal()
	if rl == nil {
		return nil, ErrPrivKeyExtr
	}

	return extractPrivKeyFromLocal(cdc, rl)
}

func extractPrivKeyFromLocal(cdc codec.Codec, rl *Record_Local) (cryptotypes.PrivKey,error) {
	if rl.PrivKeyArmor == "" {
		return nil, errors.New("private key not available")
	}

	bz := []byte(rl.PrivKeyArmor)

	switch rl.PrivKeyType {
	case "secp256k1":
		var priv secp256k1.PrivKey



		


	case "ed25519":
	}
	return priv,nil

}
