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

// TODO do we need two separate functions? does it mare sense to declare function if it is callede only one time NO two times yes?
func newLocalInfo(apk *codectypes.Any, pubKeyType string) *LocalInfo {
	return &LocalInfo{apk, pubKeyType}
}

func newLocalInfoItem(localInfo *LocalInfo) *Record_Local {
	return &Record_Local{localInfo}
}

func NewLedgerInfo(path *hd.BIP44Params) *LedgerInfo {
	return &LedgerInfo{path}
}

func NewLedgerInfoItem(ledgerInfo *LedgerInfo) *Record_Ledger {
	return &Record_Ledger{ledgerInfo}
}

func (li LedgerInfo) GetPath() *hd.BIP44Params {
	return &hd.BIP44Params{li.Path.Purpose, li.Path.CoinType, li.Path.Account, li.Path.Change, li.Path.AddressIndex}

}

func NewMultiInfo() *MultiInfo {
	return &MultiInfo{}
}

func NewMultiInfoItem(multiInfo *MultiInfo) *Record_Multi {
	return &Record_Multi{multiInfo}
}

func NewOfflineInfo() *OfflineInfo {
	return &OfflineInfo{}
}

func NewOfflineInfoItem(offlineInfo *OfflineInfo) *Record_Offline {
	return &Record_Offline{offlineInfo}
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

func (re Record) extractPrivKeyFromLocalInfo() (cryptotypes.PrivKey, error) {
	local := re.GetLocal()

	switch {
	case local != nil:
		privKey, ok := local.PrivKey.GetCachedValue().(cryptotypes.PrivKey)
		if !ok {
			return nil, fmt.Errorf("unable to cast to cryptotypes.PrivKey")
		}
		return privKey, nil
	default:
		return nil, fmt.Errorf("unable to export private rey object")
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
