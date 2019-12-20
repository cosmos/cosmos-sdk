package simapp

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

type SimappAccountCodec struct { }

func (cdc SimappAccountCodec) MarshalAccount(account exported.Account) ([]byte, error) {
	asOneof, ok := account.(isAccount_Acc)
	if !ok {
		return nil, fmt.Errorf("account %+v not handled by codec", account)
	}
	protoAcc := Account{asOneof}
	return protoAcc.Marshal()
}

func (cdc SimappAccountCodec) UnmarshalAccount(bz []byte, dest *exported.Account) error {
	var protoAcc Account
	err := protoAcc.Unmarshal(bz)
	if err != nil {
		return err
	}
	acc, ok := protoAcc.Acc.(exported.Account)
	if !ok {
		return fmt.Errorf("deserialized account %+v does not implement Account interface", acc)
	}
	*dest = acc
	return nil
}

var _ keeper.AccountCodec = SimappAccountCodec{}
