package keyring

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
)



//TODO replace Info by keyring entry in client/keys
// check  newLedgerInfo, newLocalInfo, newMultiInfo in whole codebase
// TODO count how many times newLedgerInfo or newLocalInfo is used and perhaps consider create a separate functions for that
func NewKeyringEntry(name string, pubKey *codectypes.Any, item isKeyringEntry_Item) *KeyringEntry {
	return &KeyringEntry{name, pubKey, item}
}

// TODO do we need two separate functions? does it make sense to declare function if it is callede only one time NO two times yes?
func newLocalInfo(apk *codectypes.Any, pubKeyType string) *LocalInfo {
	return &LocalInfo{apk, pubKeyType}
}

func newLocalInfoItem(localInfo *LocalInfo) *KeyringEntry_Local{
	return &KeyringEntry_Local{localInfo}
}


func newLedgerInfo(path *BIP44Params, pubKeyType string) *LedgerInfo {
	return &LedgerInfo{path, pubKeyType}
}

func newLedgerInfoItem(ledgerInfo *LedgerInfo) *KeyringEntry_Ledger{
	return &KeyringEntry_Ledger{ledgerInfo}
}

func (li LedgerInfo) GetPath() *BIP44Params {
	return li.Path
}
// TODO should I declare the function for clarity sake if it is called only once?
func newMultiInfo() *MultiInfo {
	return &MultiInfo{}
}

func newMultiInfoItem(multiInfo *MultiInfo) *KeyringEntry_Multi{
	return &KeyringEntry_Multi{multiInfo}
}

func newOfflineInfo(pubkeyType string) *OfflineInfo {
	return &OfflineInfo{pubkeyType}
}

func newOfflineInfoItem(offlineInfo *OfflineInfo) *KeyringEntry_Offline{
	return &KeyringEntry_Offline{offlineInfo}
}



func (ke KeyringEntry) GetName() string {
	return ke.Name
}

func (ke KeyringEntry) GetPubKey() (cryptotypes.PubKey, error) {
	pk, ok := ke.PubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, fmt.Errorf("Unable to cast Pubkey to cryptotypes.PubKey")
	}
	return pk, nil
}

// GetType implements Info interface
func (ke KeyringEntry) GetAddress() (types.AccAddress, error) {
	pk, err := ke.GetPubKey()
	if err != nil {
		return nil, err
	}
	return pk.Address().Bytes(), nil
}

func (ke KeyringEntry) GetAlgo() string {

	if l := ke.GetLedger(); l != nil {
		return l.PubKeyType
	}

	if o := ke.GetOffline(); o != nil {
		return o.PubKeyType
	}

	if l := ke.GetLocal(); l != nil {
		return l.PubKeyType
	}

	// there is no field pubKeyType for multi
	return ""
}

// TODO should I implement GetType?
func (ke KeyringEntry) GetType() string {
	return ""
	/*
		TODO
		fix client/keys/delete.go

		if ke.GetType() == keyring.TypeLedger || ke.GetType() == keyring.TypeOffline {
			cmd.PrintErrln("Public key reference deleted")
			continue
		}

	*/
}

func (ke KeyringEntry) extractPrivKey() (cryptotypes.PrivKey, error) {
	local := ke.GetLocal()

	switch {
	case local != nil:
		privKey, ok := local.PrivKey.GetCachedValue().(cryptotypes.PrivKey)
		if !ok {
			return nil, fmt.Errorf("unable to cast to cryptotypes.PrivKey")
		}
		return privKey, nil
	default:
		return nil, fmt.Errorf("unable to export private key object")
	}
}

// encoding info
// we remove tis function aso we can pass cdc.Marrshal install ,we put cdc on keystore
/*
func protoMarshalInfo(i Info) ([]byte, error) {
	ke, ok := i.(*KeyringEntry)
	if !ok {
		return nil, fmt.Errorf("Unable to cast Info to *KeyringEntry")
	}

	bz, err := proto.Marshal(ke)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Unable to marshal KeyringEntry to bytes")
	}

	return bz, nil
}
*/

// decoding info
// we remove tis function aso we can pass cdc.Marrshal install ,we put cdc on keystore
/*
func protoUnmarshalInfo(bz []byte, cdc codec.Codec) (Info, error) {

	var ke KeyringEntry // will not work cause we use any, use InterfaceRegistry
	// dont forget to merge master to my branch, UnmarshalBinaryBare has been renamed
	// cdcc.Marshaler.UnmarshalBinaryBare()  // like proto.UnMarshal but works with Any
	if err := cdc.UnmarshalInterface(bz, &ke); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal bytes to Info")
	}

	return ke, nil
}
*/

func NewBIP44Params(purpose uint32, coinType uint32, account uint32, change bool, adressIndex uint32) *BIP44Params {
	return &BIP44Params{purpose, coinType, account, change, adressIndex}
}

// DerivationPath returns the BIP44 fields as an array.
func (p BIP44Params) DerivationPath() []uint32 {
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


