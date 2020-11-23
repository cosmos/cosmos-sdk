# REST Endpoints Migration

Migrate your REST endpoints to the new ones in Stargate. {synopsis}

## Deprecation of Legacy REST Endpoints

The Cosmos SDK versions v0.39 and earlier provided REST endpoints to query the state and broadcast transactions. These endpoints are kept in Cosmos SDK v0.40 (Stargate), but they are marked as deprecated, and will be removed in v.41. We call these endpoints legacy REST endpoints.

Some important information concerning all legacy REST endpoints:

- Most of these endpoints are backwards-comptatible. All breaking changes are described in the next section.
- In particular, these endpoints still output Amino JSON. Cosmos v0.40 introduced Protobuf as the default encoding library throughout the codebase, but legacy REST endpoints are one of the few places in the codebase where the encoding is hardcoded to Amino. For more information about Protobuf and AMino, please read our [encoding guide](../core/encoding.md).
- All legacy REST endpoints include a [HTTP deprecation header](https://tools.ietf.org/id/draft-dalal-deprecation-header-01.html) which links to this document.

## Breaking Changes in Legacy REST Endpoints

| Legacy REST Endpoint      | Breaking Change                                                                                                                                                                                                                           |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET /staking/validators` | BondStatus is now a protobuf enum instead of an int32, and JSON serialized using its protobuf name, so expect query parameters like `?status=BOND_STATUS_{BONDED,UNBONDED,UNBONDING}` as opposed to `?status={bonded,unbonded,unbonding}` |

## Migrating to New REST Endpoints

Cosmos SDK v0.40 marks as deprecated the legacy REST endpoints, but provides for most legacy endpoint a new REST endpoint. These endpoints are automatically generated from [gRPC `Query` services](../building-modules/query-services.md) using [grpc-gateway](https://grpc-ecosystem.github.io/grpc-gateway/), so they are usually called gGPC-gateway REST endpoints.

| Legacy REST Endpoint | Description                                      | gGPC-gateway REST Endpoint         |
| -------------------- | ------------------------------------------------ | ---------------------------------- |
| `GET /txs/{hash}`    | Query tx by hash                                 | `GET /cosmos/tx/v1beta1/tx/{hash}` |
| `GET /txs`           | Query tx by events                               | `GET /cosmos/tx/v1beta1/txs`       |
| `POST /txs`          | Broadcast tx                                     | `POST /cosmos/tx/v1beta1/txs`      |
| `POST /txs/encode`   | Encodes an Amino JSON tx to an Amino binary tx   | N/A, use Protobuf directly         |
| `POST /txs/decode`   | Decodes an Amino binary tx into an Amino JSON tx | N/A, use Protobuf directly         |
