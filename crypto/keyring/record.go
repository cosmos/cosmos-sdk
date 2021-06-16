package keyring

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
)

//TODO replace Info by reyring entry in client/reys
// check  NewLedgerInfo, newLocalInfo, newMultiInfo in whole codebase
// TODO count how many times NewLedgerInfo or newLocalInfo is used and perhaps consider create a separate functions for that
func NewRecord(name string, pubKey *codectypes.Any, item isRecord_Item) *Record {
	return &Record{name, pubKey, item}
}

func newLocalRecord(apk *codectypes.Any, pubKeyType string) *Record_Local {
	return &Record_Local{apk, pubKeyType}
}

func newLocalRecordItem(localRecord *Record_Local) *Record_Local_ {
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

func (re Record) GetName() string {
	return re.Name
}

func (re Record) GetPubKey() (cryptotypes.PubKey, error) {
	pk, ok := re.PubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, fmt.Errorf("Unable to cast PubKey to cryptotypes.PubKey")
	}
	return pk, nil
}

// GetType implements Info interface
func (re Record) GetAddress() (types.AccAddress, error) {
	pk, err := re.GetPubKey()
	if err != nil {
		return nil, err
	}
	return pk.Address().Bytes(), nil
}

func (re Record) GetAlgo() string {

	if l := re.GetLocal(); l != nil {
		return l.PubKeyType
	}

	// TODO  doublecheck there is no field pubKeyType for multi,offline,ledger
	return ""
}

// TODO remove it later
func (re Record) GetType() KeyType {
	return 0
}

func (re *Record) extractPrivKeyFromLocal() (cryptotypes.PrivKey, error) {
	
	local := re.GetLocal()
	fmt.Println("extractPrivKeyFromLocal local PrivKey any", local.PrivKey)
	//"S�local PrivKey any &Any{TypeUrl:/cosmos.crypto.secp256k1.PrivKey,Value:[10 32 60 192 254 115 242 129 186 183 124 20 160 13 47 202 179 92 24 116 152 216 145 44 66 161 255 183 157 144 113 154 45 201],XXX_unrecognized:[]}"
	fmt.Println("extractPrivKeyFromLocal local PubKeyType", local.PubKeyType)
	//"secp256k1"
	
	switch {
	case local != nil:
		anyPrivKey := local.PrivKey
		privKey := anyPrivKey.GetCachedValue().(cryptotypes.PrivKey)
		fmt.Println("extractPrivKeyFromLocal privKey", privKey.String())
		/*
		fmt.Println("extractPrivKeyFromLocal ok", ok)
		if !ok {
			return nil, fmt.Errorf("unable to unpack private key")
		}
		*/
		return privKey, nil
	default:
		return nil, fmt.Errorf("unable to extract private key object")
	}
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

	var re Record // will not work cause we use any, use InterfaceRegistry
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
// TODO add tests INCORRECT LOCAL - does not include private key
func convertFromLegacyInfo(info LegacyInfo) (*Record, error) {
	fmt.Println("convertFromLegacyInfo")
	name := info.GetName()

	apk, err := codectypes.NewAnyWithValue(info.GetPubKey())
	if err != nil {
		return nil, err
	}
	
	var item isRecord_Item

	switch info.GetType() {
	case TypeLocal:
		algo := info.GetAlgo()
		localRecord := newLocalRecord(apk, string(algo))
		item = newLocalRecordItem(localRecord)
	case TypeOffline:
		emptyRecord := NewEmptyRecord()
		item = NewEmptyRecordItem(emptyRecord)
	case TypeLedger:
		path, err := info.GetPath()
		if err != nil {
			return nil, err
		}
		ledgerRecord := NewLedgerRecord(path)
		item = NewLedgerRecordItem(ledgerRecord)
	case TypeMulti:
		emptyRecord := NewEmptyRecord()
		item = NewEmptyRecordItem(emptyRecord)
	}

	kr := NewRecord(name, apk, item)
	return kr, nil
}