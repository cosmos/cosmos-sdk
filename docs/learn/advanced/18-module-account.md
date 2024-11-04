---
sidebar_position: 1
---

# Module Accounts

## Address Generation

Module accounts addresses are generated deterministically from the module name, as defined in [ADR-028](../../architecture/adr-028-public-key-addresses.md) 

Definition of account permissions is done during the app initialization

```go reference
https://github.com/cosmos/cosmos-sdk/blob/3a03804c148d0da8d6df1ad839b08c50f6896fa1/simapp/app.go#L130-L141
```

```go reference
https://github.com/cosmos/cosmos-sdk/blob/3a03804c148d0da8d6df1ad839b08c50f6896fa1/simapp/app.go#L328
```
