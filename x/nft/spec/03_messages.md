# Messages

## MsgTransferNFT

This is the most commonly expected MsgType to be supported across chains. While each application specific blockchain will have very different adoption of the `MsgMintNFT`, `MsgBurnNFT` and `MsgEditNFTMetadata` it should be expected that most chains support the ability to transfer ownership of the NFT asset. The exception to this would be non-transferable NFTs that might be attached to reputation or some asset which should not be transferable. It still makes sense for this to be represented as an NFT because there are common queriers which will remain relevant to the NFT type even if non-transferable. This Message will fail if the NFT does not exist. By default it will not fail if the transfer is executed by someone beside the owner. **It is highly recommended that a custom handler is made to restrict use of this Message type to prevent unintended use.**

| **Field** | **Type**         | **Description**                                                                                               |
|:----------|:-----------------|:--------------------------------------------------------------------------------------------------------------|
| Sender    | `sdk.AccAddress` | The account address of the user sending the NFT. By default it is __not__ required that the sender is also the owner of the NFT. |
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

This message type allows the `TokenURI` to be updated. By default anyone can execute this Message type. **It is highly recommended that a custom handler is made to restrict use of this Message type to prevent unintended use.**

| **Field**   | **Type**         | **Description**                                                                                            |
|:------------|:-----------------|:-----------------------------------------------------------------------------------------------------------|
| Sender       | `sdk.AccAddress` | The creator of the message                                      |
| ID          | `string`         | The unique ID of the NFT being edited                                                                      |
| Denom       | `string`         | The denomination of the NFT, necessary as multiple denominations are able to be represented on each chain. |
| TokenURI    | `string`         | The URI pointing to a JSON object that contains subsequent metadata information off-chain                   |

```go
// MsgEditNFTMetadata edits an NFT's metadata
type MsgEditNFTMetadata struct {
  Sender       sdk.AccAddress
  ID          string
  Denom       string
  TokenURI    string
}
```

## MsgMintNFT

This message type is used for minting new tokens. If a new `NFT` is minted under a new `Denom`, a new `Collection` will also be created, otherwise the `NFT` is added to the existing `Collection`. If a new `NFT` is minted by a new account, a new `Owner` is created, otherwise the `NFT` `ID` is added to the existing `Owner`'s `IDCollection`. By default anyone can execute this Message type. **It is highly recommended that a custom handler is made to restrict use of this Message type to prevent unintended use.**

| **Field**   | **Type**         | **Description**                                                                          |
|:------------|:-----------------|:-----------------------------------------------------------------------------------------|
| Sender      | `sdk.AccAddress` | The sender of the Message                                                                |
| Recipient   | `sdk.AccAddress` | The recipiet of the new NFT                                                              |
| ID          | `string`         | The unique ID of the NFT being minted                                                    |
| Denom       | `string`         | The denomination of the NFT.                                                             |
| TokenURI    | `string`         | The URI pointing to a JSON object that contains subsequent metadata information off-chain |

```go
// MsgMintNFT defines a MintNFT message
type MsgMintNFT struct {
  Sender      sdk.AccAddress
  Recipient   sdk.AccAddress
  ID          string
  Denom       string
  TokenURI    string
}
```

### MsgBurnNFT

This message type is used for burning tokens which destroys and deletes them. By default anyone can execute this Message type. **It is highly recommended that a custom handler is made to restrict use of this Message type to prevent unintended use.**


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
