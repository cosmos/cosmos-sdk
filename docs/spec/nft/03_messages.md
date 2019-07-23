# Messages

## MsgTransferNFT

This is the most commonly expected MsgType to be supported across chains. While each application specific blockchain will have vey different adoption of the `MsgMintNFT`, `MsgBurnNFT` and `MsgEditNFTMetadata` it should be expected that each chain supports the ability to transfer ownership of the NFT asset. Even if that transfer is heavily restricted it should be mostly supported. The exception to this would be non-transferrable NFTs that might be attached to reputation or some asset which should not be transferrable. It still makes sense for this to be represented as an NFT because there are common queriers which will remain relevant to the NFT type even if non-transferrable.

| **Field** | **Type**         | **Description**                                                                                               |
|:----------|:-----------------|:--------------------------------------------------------------------------------------------------------------|
| Sender    | `sdk.AccAddress` | The account address of the user sending the NFT. It is required that the sender is also the owner of the NFT. |
| Recipient | `sdk.AccAddress` | The account address who will receive the NFT as a result of the transfer transaction.                         |
| Denom     | `string`         | The denomination of the NFT, necessary as multiple denominations are able to be represented on each chain.    |
| ID        | `string`         | The unique ID of the NFT being transferred                                                                    |

```go
// MsgTransferNFT defines a TransferNFT message
type MsgTransferNFT struct {
  Sender    sdk.AccAddress
  Recipient sdk.AccAddress
  Denom     string
  ID        string
}
```

## MsgEditNFTMetadata

In the V1 of the NFT Module you've seen that some metadata is stored on chain along with the `TokenURI` that points to further metadata. This message type allows that specific metadata to be edited by the owner.

| **Field**   | **Type**         | **Description**                                                                                            |
|:------------|:-----------------|:-----------------------------------------------------------------------------------------------------------|
| Owner       | `sdk.AccAddress` | The owner of the NFT, which should also be the creator of the message                                      |
| ID          | `string`         | The unique ID of the NFT being edited                                                                      |
| Denom       | `string`         | The denomination of the NFT, necessary as multiple denominations are able to be represented on each chain. |
| Name        | `string`         | The name of the Token                                                                                      |
| Description | `string`         | The description of the Token                                                                               |
| Image       | `string`         | The Image of the Token                                                                                     |
| TokenURI    | `string`         | The URI pointing to a JSON object that contains subsequet metadata information off-chain                   |

```go
// MsgEditNFTMetadata edits an NFT's metadata
type MsgEditNFTMetadata struct {
  Owner       sdk.AccAddress
  ID          string
  Denom       string
  Name        string
  Description string
  Image       string
  TokenURI    string
}
```

## MsgMintNFT

This message type is used for minting new tokens. Without a restriction from a custom handler anyone can mint a new `NFT`. If a new `NFT` is minted under a new `Denom`, a new `Collection` will also be created, otherwise the `NFT` is added to the existing `Collection`. If a new `NFT` is minted by a new account, a new `Owner` is created, otherwise the `NFT` `ID` is added to the existing `Owner`.

| **Field**   | **Type**         | **Description**                                                                          |
|:------------|:-----------------|:-----------------------------------------------------------------------------------------|
| Sender      | `sdk.AccAddress` | The sender of the Message                                                                |
| Recipient   | `sdk.AccAddress` | The recipiet of the new NFT                                                              |
| ID          | `string`         | The unique ID of the NFT being minted                                                    |
| Denom       | `string`         | The denomination of the NFT.                                                             |
| Name        | `string`         | The name of the Token                                                                    |
| Description | `string`         | The description of the Token                                                             |
| Image       | `string`         | The Image of the Token                                                                   |
| TokenURI    | `string`         | The URI pointing to a JSON object that contains subsequet metadata information off-chain |

```go
// MsgMintNFT defines a MintNFT message
type MsgMintNFT struct {
  Sender      sdk.AccAddress
  Recipient   sdk.AccAddress
  ID          string
  Denom       string
  Name        string
  Description string
  Image       string
  TokenURI    string
}
```

### MsgBurnNFT

This message type is used for burning tokens which destroys and deletes them. Without a restriction from a custom handler only the owner can burn an `NFT`.

| **Field** | **Type**         | **Description**                                    |
|:----------|:-----------------|:---------------------------------------------------|
| Sender    | `sdk.AccAddress` | The account address of the user burning the token. |
| ID        | `string`         | The ID of the Token.                               |
| Denom     | `string`         | The Denom of the Token.                            |

```go
// MsgBurnNFT defines a BurnNFT message
type MsgBurnNFT struct {
  Sender sdk.AccAddress
  ID     string
  Denom  string
}
```
