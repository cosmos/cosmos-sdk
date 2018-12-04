## State

### Accounts

Accounts contain authentication information for a uniquely identified external user of an SDK blockchain,
including public key, address, and account number / sequence number for replay protection. For efficiency,
since account balances must also be fetched to pay fees, account structs also store the balance of a user
as `sdk.Coins`.

- `0x01 | Address -> amino(account)`
