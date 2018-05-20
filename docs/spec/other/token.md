# The Interchain Token Standard


The Cosmos Token Standard is proposed interface for integrators to fungible assets hosted on complaint Hubs and Zones. The Token standard is targeted at wallets, asset and portfolio  management applications, and conventional centralized digital asset exchanges that wish to monitor and transact with assets hosted on the Cosmos Hub and the connected network of Zones and other Hubs in the network.

Cosmos Zones built on top of Ethermint will follow standards designed for the Ethereum virtual machine such ERC 20.

The expected user experience for a developer integrating will be to deploy either a Light Client Daemon or a full node + LCD for the Hub or Zone chain and then integrate their systems with the REST interface.

Zones participating in this standard are expected to implement at least these methods and arguments and can implement a superset for additional functionality.


## Endpoints

### BalanceOf

**URL** : `/balanceOf`

**Method** : `GET`

**Auth required** : NO

**Data constraints**

```json
{
    "account": "[valid account encoded as hex]",
}
```

**Data example**

```json
{
    "account": "942BC57C8B9024C829627D559906D4668F4C873C",
}
```

Success Response

**Code** : `200 OK`

**Content example**

```json
[{
    "amount": 1000,
    "denom":"atom"
},
{   
    "amount":1337,
    "denom":"photon"
}]
```

Error Response

**Condition** : Bad Account Data

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "errors": [
        "Invalid Account"
    ]
}
```

### Sequence

**URL** : `/sequence`

**Method** : `GET`

**Auth required** : NO

**Data constraints**

```json
{
    "account": "[valid account encoded as hex]",
}
```

**Data example**

```json
{
    "account": "942BC57C8B9024C829627D559906D4668F4C873C",
}
```

Success Response

**Code** : `200 OK`

**Content example**

```json
{
    sequence:1,
}
```

Error Response

**Condition** : Bad Account Data

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "errors": [
        "Invalid Account"
    ]
}
```


### Transfer

This method is for creating transactions the transfer tokens between accounts.


**URL** : `/transfer`

**Method** : `POST`

**Auth required** : NO

**Signing procedure**: TODO

**Data constraints**

```json
{
    "account": "valid account encoded as hex",
    "destination_account":"valid account encoded as hex",
    "sequence": "Current sequence number for account",
    "amount": "Integer amount of tokens to send",
    "denom":"Name of token to send",
    "pubkey": "Hex encoded public key for account",
    "signature":"Signature under public key of the canoncalized data"
}
```

**Data example**

```json
{
    "account": "942BC57C8B9024C829627D559906D4668F4C873C",
    "destination_account":"942BC57C8B9024C829627D559906D4668F4C873C",
    "sequence":1,
    "amount":1000,
    "denom":"atom",
    "pubkey":"1624DE622008317FF1A0FC2EDFA2911DCBC8283D73B0F0581039D4D1FE314907EEDCF1ADF6",
    "signature":
}
```

Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "response":"Tx Ok",
}
```

Error Response

**Condition** : Invalid Signature

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "errors": [
        "Invalid Account"
    ]
}
```

Error Response

**Condition** : Invalid Signature

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "errors": [
        "Invalid Singature"
    ]
}
```

**Condition** : Insuffcient Funds

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "errors": [
        "Insufficient funds"
    ]
}
```

**Condition** : Invalid Public Key

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "errors": [
        "Invalid public keys"
    ]
}
```

### TotalSupply

**URL** : `/transfer`

**Method** : `POST`

**Auth required** : NO

**Data constraints**

**Signing procedure**: TODO

```json
{
    "account": "valid account encoded as hex",
    "sequence": "Current sequence number for account",
    "amount": "Integer amount of tokens to send",
    "denom":"Name of token to send",
    "pubkey": "Hex encoded public key for account",
    "signature":"Signature under public key of the canoncalized data"
}
```

**Data example**

```json
{
    "account": "942BC57C8B9024C829627D559906D4668F4C873C",
    "sequence":1,
    "amount":1000,
    "denom":"atom",
    "pubkey":"1624DE622008317FF1A0FC2EDFA2911DCBC8283D73B0F0581039D4D1FE314907EEDCF1ADF6",
    "signature":
}
```

Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "response":"Tx Ok",
}
```

Error Response

**Condition** : Invalid Signature

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "errors": [
        "Invalid Account"
    ]
}
```

Error Response

**Condition** : Invalid Signature

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "errors": [
        "Invalid Singature"
    ]
}
```
