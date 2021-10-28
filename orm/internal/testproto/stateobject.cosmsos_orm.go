package testproto

import (
	context "context"
	orm "github.com/cosmos/cosmos-sdk/orm/pkg/orm"
)

func NewPersonStateObjectClient(client orm.Client) PersonStateObjectClient {
	return &personStateObjectClient{client}
}

type PersonStateObjectIterator interface {
	orm.ObjectIterator
	Get() (PersonStateObject, error)
}

type PersonStateObjectClient interface {
	Create(ctx context.Context, personStateObject *PersonStateObject) error
	Get(ctx context.Context, id string) (*PersonStateObject, error)
	Update(ctx context.Context, personStateObject *PersonStateObject) error
	Delete(ctx context.Context) error
	ListByCity(ctx context.Context, city string) (PersonStateObjectIterator, error)
	ListByPostalCode(ctx context.Context, postalCode int64) (PersonStateObjectIterator, error)
	List(ctx context.Context, options orm.ListOptions) (PersonStateObjectIterator, error)
}

type personStateObjectClient struct {
	client orm.Client
}

func NewBalanceClient(client orm.Client) BalanceClient {
	return &balanceClient{client}
}

type BalanceIterator interface {
	orm.ObjectIterator
	Get() (Balance, error)
}

type BalanceClient interface {
	Create(ctx context.Context, balance *Balance) error
	Get(ctx context.Context, address string, denom string) (*Balance, error)
	Update(ctx context.Context, balance *Balance) error
	Delete(ctx context.Context) error
	ListByAddress(ctx context.Context, address string) (BalanceIterator, error)
	ListByDenom(ctx context.Context, denom string) (BalanceIterator, error)
	List(ctx context.Context, options orm.ListOptions) (BalanceIterator, error)
}

type balanceClient struct {
	client orm.Client
}

func NewParamClient(client orm.Client) ParamClient {
	return &paramClient{client}
}

type ParamClient interface {
	Get(ctx context.Context) (*Param, error)
	Create(ctx context.Context, param *Param) error
	Update(ctx context.Context, param *Param) error
	Delete(ctx context.Context) error
}

type paramClient struct {
	client orm.Client
}

func NewRedelegationClient(client orm.Client) RedelegationClient {
	return &redelegationClient{client}
}

type RedelegationIterator interface {
	orm.ObjectIterator
	Get() (Redelegation, error)
}

type RedelegationClient interface {
	Create(ctx context.Context, redelegation *Redelegation) error
	Get(ctx context.Context, delegatorSrcAddress string, validatorSrcAddress string, validatorDstAddress string) (*Redelegation, error)
	Update(ctx context.Context, redelegation *Redelegation) error
	Delete(ctx context.Context) error
	ListByDelegatorSrcAddress(ctx context.Context, delegatorSrcAddress string) (RedelegationIterator, error)
	ListByValidatorSrcAddress(ctx context.Context, validatorSrcAddress string) (RedelegationIterator, error)
	ListByValidatorDstAddress(ctx context.Context, validatorDstAddress string) (RedelegationIterator, error)
	List(ctx context.Context, options orm.ListOptions) (RedelegationIterator, error)
}

type redelegationClient struct {
	client orm.Client
}
