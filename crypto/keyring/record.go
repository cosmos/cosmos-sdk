package keyring

import (
	"errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
)

var ErrPrivKeyExtr = errors.New("private key extraction works only for Local")

func NewRecord(name string, pk cryptotypes.PubKey, item isRecord_Item) (*Record, error) {
	any, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return nil, err
	}

	return &Record{name, any, item}, nil
}

func NewLocalRecord(priv cryptotypes.PrivKey) (*Record_Local, error) {
	any, err := codectypes.NewAnyWithValue(priv)
	if err != nil {
		return nil, err
	}

	return &Record_Local{any, priv.Type()}, nil
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

func NewOfflineRecord() *Record_Offline {
	return &Record_Offline{}
}

func NewOfflineRecordItem(re *Record_Offline) *Record_Offline_ {
	return &Record_Offline_{re}
}

func NewMultiRecord() *Record_Multi {
	return &Record_Multi{}
}

func NewMultiRecordItem(re *Record_Multi) *Record_Multi_ {
	return &Record_Multi_{re}
}

func (k *Record) GetPubKey() (cryptotypes.PubKey, error) {
	pk, ok := k.PubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errors.New("unable to cast any to cryptotypes.PubKey")
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

func (k Record) GetType() KeyType {
	switch {
	case k.GetLocal() != nil:
		return TypeLocal
	case k.GetLedger() != nil:
		return TypeLedger

	case k.GetMulti() != nil:
		return TypeMulti
		// k.Getoffline()
	default:
		return TypeOffline
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (k *Record) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pk cryptotypes.PubKey
	if err := unpacker.UnpackAny(k.PubKey, &pk); err != nil {
		return err
	}

	if l := k.GetLocal(); l != nil {
		var priv cryptotypes.PrivKey
		return unpacker.UnpackAny(l.PrivKey, &priv)
	}

	return nil
}

func ExtractPrivKeyFromRecord(k *Record) (cryptotypes.PrivKey, error) {
	rl := k.GetLocal()
	if rl == nil {
		return nil, ErrPrivKeyExtr
	}

	return extractPrivKeyFromLocal(rl)
}

func extractPrivKeyFromLocal(rl *Record_Local) (cryptotypes.PrivKey, error) {
	if rl.PrivKey == nil {
		return nil, errors.New("private key is not available")
	}

	priv, ok := rl.PrivKey.GetCachedValue().(cryptotypes.PrivKey)
	if !ok {
		return nil, errors.New("unable to cast any to cryptotypes.PrivKey")
	}

	return priv, nil
}
