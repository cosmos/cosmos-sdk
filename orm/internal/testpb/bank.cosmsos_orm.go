package testpb

import (
	context "context"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"

	ormtable "github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

type BankStoreReadAccessor interface {
	OpenRead(context.Context) BankReader
}

type BankStoreAccessor interface {
	Open(context.Context) BankStore
}

func NewBankStoreAccessor() (BankStoreAccessor, error) {
	panic("TODO")
}

type BankStore interface {
	BalanceStore
	SupplyStore
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
	ListBalance(BalanceIndexKey) (BalanceIterator, error)
}

type BalanceIterator struct {
	ormtable.Iterator
}

type BalanceIndexKey interface {
	id() uint32
	values() []protoreflect.Value
	balanceIndexKey()
}

type BalanceAddressDenomIndexKey struct {
}

type BalanceDenomAddressIndexKey struct {
}

type SupplyStore interface {
	SupplyReader

	CreateSupply(supply *Supply) error
	UpdateSupply(supply *Supply) error
	SaveSupply(supply *Supply) error
	DeleteSupply(supply *Supply) error
}

type SupplyReader interface {
	HasSupply(denom string) (found bool, err error)
	GetSupply(denom string) (*Supply, error)
	ListSupply(SupplyIndexKey) (SupplyIterator, error)
}

type SupplyIterator struct {
	ormtable.Iterator
}

type SupplyIndexKey interface {
	id() uint32
	values() []protoreflect.Value
	supplyIndexKey()
}

type SupplyDenomIndexKey struct {
}

type bankStoreAcc struct {
	balanceTable ormtable.Table
	supplyTable  ormtable.Table
}

type bankStore struct {
	balanceTable ormtable.Table
	supplyTable  ormtable.Table
}
