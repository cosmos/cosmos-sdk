# The Interchain Token Standard

The Cosmos Interchain Token Standard is the proposed interface to allow the transfer of fungible 
assets hosted on hubs and zones within the Cosmos ecosystem. The goal of the token standard is to
allow client implementors, such as wallet applications, exchanges or monitoring tools to interact 
with assets (fungible tokens) that are hosted on the hubs and connected zones.

The ITS (Interchain Token Standard) supports an arbitrary number of assets per account.

Ethermint implements the standard ERC20 interface but the Interchain Token Standard provides an 
adapter for it.

A developer integrating their zone with the ITS (Interchain Token Standard) involves deploying a 
light client daemon (LCD) or a full node + LCD for the given hub or zone. Then any client 
implementor (wallet, exchange, ...) is able to operate with the newly created and deployed zone.

Zones participating in this standard are expected to implement at the following methods and 
arguments.

## Endpoints

### Balance

**URL** : `/balance`

**Method** : `GET`

**Arguments**

```json
{
    "account": "[valid account encoded as hex]",
}
```

**Argument Example**

```json
{
    "account": "942BC57C8B9024C829627D559906D4668F4C873C",
}
```

Success Response

**Code** : `200 OK`

**Response Example**

```json
{
    "atom": 1000,
    "photon": 500,
    "ether": 20
}
```

Error Response

**Condition** : Bad Account Data

**Code** : `400 BAD REQUEST`

**Response Example** :

```json
{
    "errors": [
        "Invalid Account"
    ]
}
```

### Create Transfer

This method is for creating transactions the transfer tokens between accounts.

**URL** : `/create_transfer`

**Method** : `GET`

**Arguments**

```json
{
    "from": "valid account encoded as hex",
    "to":"valid account encoded as hex",
    "amount": "Integer amount of tokens to send",
    "denom":"Name of token to send",
}
```

**Argument Example**

```json
{
    "from": "942BC57C8B9024C829627D559906D4668F4C873C",
    "to":"942BC57C8B9024C829627D559906D4668F4C873C",
    "amount":1000,
    "denom":"atom",
}
```

Success Response

**Code** : `200 OK`

**Response Example**

```json
{
    "transaction": "[]bytes of a valid transaction bytes to be signed for that zone"
}
```

Error Response

**Condition** : Invalid Arguments

**Code** : `400 BAD REQUEST`

**Response Example** :

```json
{
    "errors": [
        "Invalid Account",
        "Insufficient Funds",
    ]
}
```

### Signed Transfer

This method is for sending a signed transactions to transfer tokens.

**URL** : `/signed_transfer`

**Method** : `POST`

**Arguments**

```json
{
    "signed_transfer": "[]bytes of a valid transaction and it's signature"
}
```

**Argument Example**

```json
{
    "signed_transfer": "TODO"
}
```

Success Response

**Code** : `200 OK`

**Response Example**

```json
{
    "tx_hash": "TODO"
}
```

Error Response

**Condition** : Invalid Arguments

**Code** : `400 BAD REQUEST`

**Response Example** :

```json
{
    "errors": [
        "Invalid Signature"
    ]
}
```
