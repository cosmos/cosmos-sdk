package bankexample

import (
	context "context"

	"github.com/cosmos/cosmos-sdk/orm/model/ormlist"

	ormtable "github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

type StateStoreAccessor interface {
	StateReaderAccessor

	Open(context.Context) StateStore
}

type StateReaderAccessor interface {
	OpenRead(context.Context) StateReader
}

func NewStateStoreAccessor() (StateStoreAccessor, error) {
	panic("TODO")
}

func NewStateReaderAccessor() (StateReaderAccessor, error) {
	panic("TODO")
}

type StateStore interface {
	StateReader

	BalanceStore
	SupplyStore
}

type StateReader interface {
	BalanceReader
	SupplyReader
}

type BalanceStore interface {
	BalanceReader

	CreateBalance(balance *Balance) error
	UpdateBalance(balance *Balance) error
	SaveBalance(balance *Balance) error
	DeleteBalance(balance *Balance) error
}

type BalanceReader interface {
	HasBalance(address string, denom string) (found bool, err error)
	GetBalance(address string, denom string) (*Balance, error)
	ListBalance(BalanceIndexKey, ...ormlist.Option) (BalanceIterator, error)
	ListBalanceRange(from, to BalanceIndexKey, opts ...ormlist.Option) (BalanceIterator, error)
}

type BalanceIterator struct {
	ormtable.Iterator
}

func (i BalanceIterator) Value() (*Balance, error) {
	balance := &Balance{}
	err := i.UnmarshalMessage(balance)
	return balance, err
}

type BalanceIndexKey interface {
	id() uint32
	values() []interface{}
	balanceIndexKey()
}

type BalanceAddressDenomIndexKey struct {
}

type BalanceDenomAddressIndexKey struct {
}

type SupplyReader interface {
	HasSupply(denom string) (found bool, err error)
	GetSupply(denom string) (*Supply, error)
	ListSupply(SupplyIndexKey, ...ormlist.Option) (SupplyIterator, error)
	ListSupplyRange(from, to SupplyIndexKey, opts ...ormlist.Option) (SupplyIterator, error)
}

type SupplyStore interface {
	SupplyReader

	CreateSupply(supply *Supply) error
	UpdateSupply(supply *Supply) error
	SaveSupply(supply *Supply) error
	DeleteSupply(supply *Supply) error
}

type SupplyIterator interface {
	Next() bool
	Value() (*Supply, error)
}

type SupplyIndexKey interface {
	id() uint32
	values() []interface{}
	supplyIndexKey()
}

type SupplyDenomIndexKey struct {
}

var _ StateReader = &stateReader{}

type stateReader struct {
	*stateStoreAccessor
	ctx context.Context
}

func (s stateReader) HasBalance(address string, denom string) (found bool, err error) {
	return s.balanceTable.PrimaryKey().Has(s.ctx, address, denom)
}

func (s stateReader) GetBalance(address string, denom string) (*Balance, error) {
	balance := &Balance{}
	found, err := s.balanceTable.PrimaryKey().Get(s.ctx, balance, address, denom)
	if !found {
		return nil, nil
	}
	return balance, err
}

func (s stateReader) ListBalance(key BalanceIndexKey, opts ...ormlist.Option) (BalanceIterator, error) {
	opts = append(opts, ormlist.Prefix(key.values()...))
	it, err := s.balanceTable.Iterator(s.ctx, opts...)
	return BalanceIterator{it}, err
}

func (s stateReader) ListBalanceRange(from, to BalanceIndexKey, opts ...ormlist.Option) (BalanceIterator, error) {
	//TODO implement me
	panic("implement me")
}

func (s stateReader) HasSupply(denom string) (found bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (s stateReader) GetSupply(denom string) (*Supply, error) {
	//TODO implement me
	panic("implement me")
}

func (s stateReader) ListSupply(key SupplyIndexKey, option ...ormlist.Option) (SupplyIterator, error) {
	//TODO implement me
	panic("implement me")
}

func (s stateReader) ListSupplyRange(from, to SupplyIndexKey, opts ...ormlist.Option) (SupplyIterator, error) {
	//TODO implement me
	panic("implement me")
}

type stateStore struct {
	stateReader
}

type stateStoreAccessor struct {
	balanceTable ormtable.Table
	supplyTable  ormtable.Table
}
