# Events

The nft module emits the following events:

## Handlers

### MsgTransferNFT

| Type         | Attribute Key | Attribute Value    |
|--------------|---------------|--------------------|
| transfer_nft | denom         | {nftDenom}         |
| transfer_nft | nft-id        | {nftID}            |
| transfer_nft | recipient     | {recipientAddress} |
| message      | module        | nft                |
| message      | action        | transfer_nft       |
| message      | sender        | {senderAddress}    |

### MsgEditNFTMetadata

| Type              | Attribute Key | Attribute Value   |
|-------------------|---------------|-------------------|
| edit_nft_metadata | denom         | {nftDenom}        |
| edit_nft_metadata | nft-id        | {nftID}           |
| message           | module        | nft               |
| message           | action        | edit_nft_metadata |
| message           | sender        | {senderAddress}   |
| message           | token-uri     | {tokenURI}        |

### MsgMintNFT

| Type     | Attribute Key | Attribute Value |
|----------|---------------|-----------------|
| mint_nft | denom         | {nftDenom}      |
| mint_nft | nft-id        | {nftID}         |
| message  | module        | nft             |
| message  | action        | mint_nft        |
| message  | sender        | {senderAddress} |
| message  | token-uri     | {tokenURI}      |

### MsgBurnNFTs

| Type     | Attribute Key | Attribute Value |
|----------|---------------|-----------------|
| burn_nft | denom         | {nftDenom}      |
| burn_nft | nft-id        | {nftID}         |
| message  | module        | nft             |
| message  | action        | burn_nft        |
| message  | sender        | {senderAddress} |
