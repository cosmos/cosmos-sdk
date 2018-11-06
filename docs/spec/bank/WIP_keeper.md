WORK IN PROGRESS 
See PR comments here https://github.com/cosmos/cosmos-sdk/pull/2072

# Keeper

## Denom Metadata

The BankKeeper contains a store that stores the metadata of different token denoms.  Denoms are referred to by their name, same as the `denom` field in sdk.Coin.  The different attributes of a denom are stored in the denom metadata store under the key `[denom name]:[attribute name]`.  The default attributes in the store are explained below.  However, this can be extended by the developer or through SoftwareUpgrade proposals.

### Decimals `int8`

- `Base Unit` = The common standard for the default "standard" size of a token. Examples: 1 Bitcoin or 1 Ether.
- `Smallest Unit` = The smallest possible denomination of a token. A fraction of the base unit.  Examples: 1 satoshi or 1 wei.

All amounts throughout the SDK are denominated in the smallest unit of a token, so that all amounts can be expressed as integers.  However, UIs typically want to display token values in the base unit, so the Decimals metadata field standardizes the number of digits that come after the decimal place in the base unit.

`1 [Base Unit] = 10^(N) [Smallest Unit]`

### TotalSupply `sdk.Integer`

The TotalSupply of a denom is the total amount of a token that exists (known to the chain) across all accounts and modules.  It is denominated in the `smallest unit` of a denom.  It can be changed by the Keeper functions `MintCoins` and `BurnCoins`.  `AddCoins` and `SubtractCoins` are used when adding or subtracting coins for an account, but not removing them from total supply (for example, when moving the coins to the control of the staking module).

### Aliases `[]string`

Aliases is an array of strings that are "alternative names" for a token. As an example, while the Ether's denom name might be `ether`, a possible alias could be `ETH`.  This field can be useful for UIs and clients.  It is intended that this field can be modified by a governance mechanism.
