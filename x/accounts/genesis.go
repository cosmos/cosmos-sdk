package accounts

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	v1 "cosmossdk.io/x/accounts/v1"
)

func (k Keeper) ExportState(ctx context.Context) (*v1.GenesisState, error) {
	genState := &v1.GenesisState{}

	// get account number
	accountNumber, err := k.AccountNumber.Peek(ctx)
	if err != nil {
		return nil, err
	}

	genState.AccountNumber = accountNumber

	err = k.AccountsByType.Walk(ctx, nil, func(accAddr []byte, accType string) (stop bool, err error) {
		accNum, err := k.AccountByNumber.Get(ctx, accAddr)
		if err != nil {
			return true, err
		}
		accState, err := k.exportAccount(ctx, accAddr, accType, accNum)
		if err != nil {
			return true, err
		}
		genState.Accounts = append(genState.Accounts, accState)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return genState, nil
}

func (k Keeper) exportAccount(ctx context.Context, addr []byte, accType string, accNum uint64) (*v1.GenesisAccount, error) {
	addrString, err := k.addressCodec.BytesToString(addr)
	if err != nil {
		return nil, err
	}
	account := &v1.GenesisAccount{
		Address:       addrString,
		AccountType:   accType,
		AccountNumber: accNum,
		State:         nil,
	}
	rng := collections.NewPrefixedPairRange[uint64, []byte](accNum)
	err = k.AccountsState.Walk(ctx, rng, func(key collections.Pair[uint64, []byte], value []byte) (stop bool, err error) {
		account.State = append(account.State, &v1.KVPair{
			Key:   key.K2(),
			Value: value,
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (k Keeper) ImportState(ctx context.Context, genState *v1.GenesisState) error {
	var largestNum *uint64
	var err error
	// import accounts
	for _, acc := range genState.Accounts {
		err = k.importAccount(ctx, acc)
		if err != nil {
			return fmt.Errorf("%w: %s", err, acc.Address)
		}

		accNum := acc.AccountNumber

		if largestNum == nil || *largestNum < accNum {
			largestNum = &accNum
		}
	}

	if largestNum != nil {
		// set the account number to the highest account number to avoid duplicate account number
		err = k.AccountNumber.Set(ctx, *largestNum)
	}

	return err
}

func (k Keeper) importAccount(ctx context.Context, acc *v1.GenesisAccount) error {
	// Check if the account type exists in the registered accounts
	_, ok := k.accounts[acc.AccountType]
	if !ok {
		// If the account type does not exist, return an error
		return fmt.Errorf("account type %s not found in the registered accounts", acc.AccountType)
	}

	addrBytes, err := k.addressCodec.StringToBytes(acc.Address)
	if err != nil {
		return err
	}
	err = k.AccountsByType.Set(ctx, addrBytes, acc.AccountType)
	if err != nil {
		return err
	}
	err = k.AccountByNumber.Set(ctx, addrBytes, acc.AccountNumber)
	if err != nil {
		return err
	}
	for _, kv := range acc.State {
		err = k.AccountsState.Set(ctx, collections.Join(acc.AccountNumber, kv.Key), kv.Value)
		if err != nil {
			return err
		}
	}
	return nil
}
