package bankexample

import (
	context "context"

	ormtable "github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

type StateStoreAccessor interface {
	Open(context.Context) StateStore
}

func NewStateStoreAccessor() (StateStoreAccessor, error) {
}

type StateStore interface {
	BalanceStore
	SupplyStore
}

type BalanceStore interface {
	CreateBalance(balance *Balance) error
	UpdateBalance(balance *Balance) error
	SaveBalance(balance *Balance) error
	DeleteBalance(balance *Balance) error
}

type SupplyStore interface {
	CreateSupply(supply *Supply) error
	UpdateSupply(supply *Supply) error
	SaveSupply(supply *Supply) error
	DeleteSupply(supply *Supply) error
}

type stateStore struct {
	balanceTable ormtable.Table
	supplyTable  ormtable.Table
}
