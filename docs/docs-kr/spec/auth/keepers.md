## 키퍼

auth 모듈은 하나의 계정 키퍼(account keeper) 하나만을 이용합니다. 해당 키퍼는 계정을 읽고 쓰는데 이용됩니다.

### 계정 키퍼

현재로써는 모든 계정에 읽기 쓰기가 가능한 키퍼는 하나만 있는 상태입니다.

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
