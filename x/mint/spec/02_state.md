<!--
order: 2
-->

# State

## Minter

The minter is a space for holding current inflation information.

 - Minter: `0x00 -> ProtocolBuffer(minter)`

```protobuf
message Minter {
  // current annual inflation rate
  string inflation = 1 [(gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec"];
  // current annual expected provisions
  string annual_provisions = 2 [(gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec"];
}
```

## Params

Minting params are held in the global params store. 

 - Params: `mint/params -> ProtocolBuffer(params)`

```protobuf
message Params {
	// type of coin to mint
	string mint_denom = 1;
	// maximum annual change in inflation rate
  string inflation_rate_change = 2 [(gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec"];
  // maximum inflation rate
  string inflation_max = 3 [(gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec"];
  // minimum inflation rate
  string inflation_min = 4 [(gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec"];
  // goal of percent bonded atoms
  string goal_bonded = 5 [(gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec"];
  // expected blocks per year
  uint64 blocks_per_year = 6;
}
```
