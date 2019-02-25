# Keepers

The auth module only exposes one keeper, the account keeper, which can be used to read and write accounts.

## Account Keeper

Presently only one fully-permissioned account keeper is exposed, which has the ability to both read and write
all fields of all accounts, and to iterate over all stored accounts.

```golang
type AccountKeeper interface {
  // Return a new account with the next account number and the specified address. Does not save the new account to the store.
  NewAccountWithAddress(AccAddress) Account

  // Return a new account with the next account number. Does not save the new account to the store.
  NewAccount(Account) Account

  // Retrieve an account from the store
  GetAccount(AccAddress) Account

  // Set an account in the store
  SetAccount(Account)

  // Remove an account from the store
  RemoveAccount(Account)

  // Iterate over all accounts, calling the provided function. Stop iteraiton when it returns false.
  IterateAccounts(func(Account) (bool))

  // Fetch the public key of an account at a specified address
  GetPubKey(AccAddress) PubKey

  // Fetch the sequence of an account at a specified address
  GetSequence(AccAddress) uint64

  // Fetch the next account number, and increment the internal counter
  GetNextAccountNumber() uint64
}
```
