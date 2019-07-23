# NFT Specification

## Overview

The NFT Module described here is meant to be used as a module across chains for managing non-fungible token that represent individual assets with unique features. This standard was first developed on Ethereum within the ERC-721 and the subsequent EIP of the same name. This standard utilized the features of the Ethereum blockchain as well as the restrictions. The subsequent ERC-1155 standard addressed some of the restrictions of Ethereum regarding storage costs and semi-fungible assets.

NFTs on application specific blockchains share some but not all features as their Ethereum brethren. Since application specific blockchains are more flexible in how their resources are utilized it makes sense that should have the option of exploiting those resources. This includes the aility to use strings as IDs and to optionally store metadata on chain. The user-flow of composability with smart contracts should also be rethought on application specific blockchains with regard to Inter-Blockchain Communication as it is a different design experience from communication between smart contracts.

## Contents

1. **[Concepts](./01_concepts.md)**
  - [NFT](./01_concepts.md#nft)
  - [Collections](./01_concepts.md#collections)
2. **[State](./02_state.md)**
	- [Collections](./02_state.md#collections)
  - [Owners](./02_state.md#owners)
3. **[Messages](./03_messages.md)**
	- [Transfer NFT](./03_messages.md#transfer-nft)
	- [Edit Metadata](./03_messages.md#edit-metadata)
  - [Mint NFT](./03_messages.md#mint-nft)
	- [Burn NFT](./03_messages.md#burn-nft)
4. **[Events](./04_events.md)**
5. **[Future Improvements](./05_future_improvements.md)**

## 