# Auth module specification

## Abstract

This document specifies the auth module of the Cosmos SDK.

The auth module is responsible for specifying the base transaction and account types
for an application, since the SDK itself is agnostic to these particulars. It contains
the ante handler, where all basic transaction validity checks (signatures, nonces, auxiliary fields)
are performed, and exposes the account keeper, which allows other modules to read, write, and modify accounts.

This module will be used in the Cosmos Hub.

## Contents

1. **[State](state.md)**
    1. [Accounts](state.md#accounts)
        1. [Account Interface](state.md#account-interface)
        2. [Base Account](state.md#baseaccount)
        3. [Vesting Account](state.md#vestingaccount)
1. **[Types](types.md)**
    1. [StdFee](types.md#stdfee)
    2. [StdSignature](types.md#stdsignature)
    3. [StdTx](types.md#stdtx)
    4. [StdSignDoc](types.md#stdsigndoc)
1. **[Keepers](keepers.md)**
    1. [AccountKeeper](keepers.md#account-keeper)
1. **[Handlers](handlers.md)**
    1. [Ante Handler](handlers.md#ante-handler)
1. **[Gas & Fees](gas_fees.md)**
