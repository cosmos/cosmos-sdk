---
sidebar_position: 1
---

# Module Accounts

## Address Generation

Module account addresses are generated deterministically from the module name as defined in [ADR-028](https://docs.cosmos.network/v0.52/build/architecture/adr-028-public-key-addresses#module-account-addresses)

Definition of account permissions is done during the app initialization

```go reference
https://github.com/cosmos/cosmos-sdk/blob/3a03804c148d0da8d6df1ad839b08c50f6896fa1/simapp/app.go#L130-L141
```

```go reference
https://github.com/cosmos/cosmos-sdk/blob/3a03804c148d0da8d6df1ad839b08c50f6896fa1/simapp/app.go#L328
```
