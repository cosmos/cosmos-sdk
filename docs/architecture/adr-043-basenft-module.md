# ADR 43: BaseNFT Module

## Changelog

- 05.05.2021: Initial Draft

## Status

DRAFT

## Abstract

This ADR defines a generic NFT module of `x/nft` which supports NFTs as a `proto.Any` and contains `BaseNFT` as the default implementation.

## Context

As was discussed in [#9065](https://github.com/cosmos/cosmos-sdk/discussions/9065), several potential solutions can be considered:
- irismod/nft and modules/incubator/nft
- CW721
- DID NFTs
- interNFT

Considering generic usage and compatibility of interchain protocols including IBC and Gravity Bridge, it is preferred to have a very simple NFT module design which only stores NFTs by id and owner. 

Application-specific functions (minting, burning, transferring, etc.) should be managed by other modules on Cosmos Hub or on other Zones.

The current design is based on the work done by [IRISnet team](https://github.com/irisnet/irismod/tree/master/modules/nft) and an older implementation in the [Cosmos repository](https://github.com/cosmos/modules/tree/master/incubator/nft).


## Decision

We will create a module `x/nft` which only stores NFTs by id and owner.

The interface for the `x/nft` module:

```go
// NFTI is an interface used to store NFTs at a given id and owner.
type NFTI interface {
	GetId() string // can not return empty string.
	GetOwner() sdk.AccAddress
}
```

We will also create `BaseNFT` as the default implementation of the `NFTI` interface:
```proto
message BaseNFT {
  option (gogoproto.equal) = true;

  string id    = 1;
  string name  = 2;
  string uri   = 3 [(gogoproto.customname) = "URI"];
  string data  = 4;
  string owner = 5;
}
```

Queries functions for `BaseNFT` is:
```proto
service Query {

  // NFT queries NFT details based on id.
  rpc NFT(QueryNFTRequest) returns (QueryNFTResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/nfts/{id}";
  }

  // NFTs queries all proposals based on the optional onwer.
  rpc NFTs(QueryNFTsRequest) returns (QueryNFTsResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/nfts";
  }
}

// QueryNFTRequest is the request type for the Query/NFT RPC method
message QueryNFTRequest {
  string id = 1;
}

// QueryNFTResponse is the response type for the Query/NFT RPC method
message QueryNFTResponse {
  google.protobuf.Any nft = 1 [(cosmos_proto.accepts_interface) = "NFTI", (gogoproto.customname) = "NFT"];
}

// QueryNFTsRequest is the request type for the Query/NFTs RPC method
message QueryNFTsRequest {
  string                                 owner      = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}

// QueryNFTsResponse is the response type for the Query/NFTs RPC method
message QueryNFTsResponse {
  repeated google.protobuf.Any nfts = 1 [(cosmos_proto.accepts_interface) = "NFTI", (gogoproto.customname) = "NFTs"];
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```



## Consequences

### Backward Compatibility

No backwards incompatibilities.

### Positive

- NFT functions available on Cosmos Hub
- NFT module which supports interoperability with IBC and other cross-chain infrastractures like Gravity Bridge

### Negative

### Neutral

- Other functions need more modules. For example, a custody module is needed for NFT trading function, a collectible module is needed for defining NFT properties

## Further Discussions

For other kinds of applications on the Hub, more app-specific modules can be developed in the future:
- `x/collectibles`: grouping NFTs into collections and defining properties of NFTs such as minting, burning and transferring, etc.
- `x/nft_custody`: custody of NFTs to support trading functionality
- `x/nft_marketplace`: selling and buying NFTs using sdk.Coins

Other networks in the Cosmos ecosystem could design and implement their own NFT modules for specific NFT applications and usecases.

## References

- Initial discussion: https://github.com/cosmos/cosmos-sdk/discussions/9065
- x/nft: initialize module: https://github.com/cosmos/cosmos-sdk/pull/9174