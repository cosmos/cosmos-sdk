# REST Endpoints Migration

Migrate your REST endpoints to the Stargate ones. {synopsis}

## Deprecation of Legacy REST Endpoints

The Cosmos SDK versions v0.39 and earlier provided REST endpoints to query the state and broadcast transactions. These endpoints are kept in Cosmos SDK v0.40 (Stargate), but they are marked as deprecated, and will be removed in v.41. We call these endpoints legacy REST endpoints.

Some important information concerning all legacy REST endpoints:

- Most of these endpoints are backwards-comptatible. All breaking changes are described in the next section.
- In particular, these endpoints still output Amino JSON. Cosmos v0.40 introduced Protobuf as the default encoding library throughout the codebase, but legacy REST endpoints are one of the few places in the codebase where the encoding is hardcoded to Amino. For more information about Protobuf and AMino, please read our [encoding guide](../core/encoding.md).
- All legacy REST endpoints include a [HTTP deprecation header](https://tools.ietf.org/id/draft-dalal-deprecation-header-01.html) which links to this document.

## Breaking Changes in Legacy REST Endpoints

| Legacy REST Endpoint      | Description        | Breaking Change                                                                                                                                                                                                                            |
| ------------------------- | ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GET /txs/{hash}`         | Query tx by hash   | Endpoint will error when trying to output non-Amino txs (e.g. IBC txs).                                                                                                                                                                    |
| `GET /txs`                | Query tx by events | Endpoint will error when trying to output non-Amino txs (e.g. IBC txs).                                                                                                                                                                    |
| `GET /staking/validators` | Get all validators | BondStatus is now a protobuf enum instead of an int32, and JSON serialized using its protobuf name, so expect query parameters like `?status=BOND_STATUS_{BONDED,UNBONDED,UNBONDING}` as opposed to `?status={bonded,unbonded,unbonding}`. |

## Migrating to New REST Endpoints

Cosmos SDK v0.40 marks as deprecated the legacy REST endpoints, but provides for most legacy endpoint a new REST endpoint. These endpoints are automatically generated from [gRPC `Query` services](../building-modules/query-services.md) using [grpc-gateway](https://grpc-ecosystem.github.io/grpc-gateway/), so they are usually called gGPC-gateway REST endpoints.

| Legacy REST Endpoint                                                   | Description                                                    | New gGPC-gateway REST Endpoint                                                                        |
| ---------------------------------------------------------------------- | -------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `GET /txs/{hash}`                                                      | Query tx by hash                                               | `GET /cosmos/tx/v1beta1/tx/{hash}`                                                                    |
| `GET /txs`                                                             | Query tx by events                                             | `GET /cosmos/tx/v1beta1/txs`                                                                          |
| `POST /txs`                                                            | Broadcast tx                                                   | `POST /cosmos/tx/v1beta1/txs`                                                                         |
| `POST /txs/encode`                                                     | Encodes an Amino JSON tx to an Amino binary tx                 | N/A, use Protobuf directly                                                                            |
| `POST /txs/decode`                                                     | Decodes an Amino binary tx into an Amino JSON tx               | N/A, use Protobuf directly                                                                            |
| `POST /bank/*`                                                         | Create unsigned `Msg`s for bank tx                             | N/A, use Protobuf directly                                                                            |
| `GET /bank/balances/{address}`                                         | Get the balance of an address                                  | `GET /cosmos/bank/v1beta1/balances/{address}/{denom}`                                                 |
| `GET /bank/total`                                                      | Get the total supply of all coins                              | `GET /cosmos/bank/v1beta1/supply`                                                                     |
| `GET /bank/total/{denom}`                                              | Get the total supply of one coin                               | `GET /cosmos/bank/v1beta1/supply/{denom}`                                                             |
| `POST /distribution/delegators/{delegatorAddr}/rewards`                | Withdraw all delegator rewards                                 | N/A, use Protobuf directly                                                                            |
| `POST /distribution/*`                                                 | Create unsigned `Msg`s for distribution                        | N/A, use Protobuf directly                                                                            |
| `GET /distribution/delegators/{delegatorAddr}/rewards`                 | Get the total rewards balance from all delegations             | `GET /cosmos/distribution/v1beta1/v1beta1/delegators/{delegator_address}/rewards`                     |
| `GET /distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}` | Query a delegation reward                                      | `GET /cosmos/distribution/v1beta1/delegators/{delegatorAddr}/rewards/{validatorAddr}`                 |
| `GET /distribution/delegators/{delegatorAddr}/withdraw_address`        | Get the rewards withdrawal address                             | `GET /cosmos/distribution/v1beta1/delegators/{delegatorAddr}/withdraw_address`                        |
| `GET /distribution/validators/{validatorAddr}`                         | Validator distribution information                             | `GET /cosmos/distribution/v1beta1/validators/{validatorAddr}`                                         |
| `GET /distribution/validators/{validatorAddr}/rewards`                 | Commission and self-delegation rewards of a single a validator | `GET /cosmos/distribution/v1beta1/validators/{validatorAddr}/rewards`                                 |
| `GET /distribution/validators/{validatorAddr}/outstanding_rewards`     | Outstanding rewards of a single validator                      | `GET /cosmos/distribution/v1beta1/validators/{validatorAddr}/outstanding_rewards`                     |
| `GET /distribution/parameters`                                         | Get the current distribution parameter values                  | `GET /cosmos/distribution/v1beta1/params`                                                             |
| `GET /distribution/community_pool`                                     | Get the amount held in the community pool                      | `GET /cosmos/distribution/v1beta1/community_pool`                                                     |
| `GET /evidence/{evidence-hash}`                                        | Get evidence by hash                                           | `GET /cosmos/evidence/v1beta1/evidence/{evidence_hash}`                                               |
| `GET /evidence`                                                        | Get all evidence                                               | `GET /cosmos/evidence/v1beta1/evidence`                                                               |
| `POST /gov/*`                                                          | Create unsigned `Msg`s for gov                                 | N/A, use Protobuf directly                                                                            |
| `GET /gov/parameters/{type}`                                           | Get government parameters                                      | `GET /cosmos/gov/params/{type}`                                                                       |
| `GET /gov/proposals`                                                   | Get all proposals                                              | `GET /cosmos/gov/proposals`                                                                           |
| `GET /gov/proposals/{proposal-id}`                                     | Get proposal by id                                             | `GET /cosmos/gov/proposals/{proposal-id}`                                                             |
| `GET /gov/proposals/{proposal-id}/proposer`                            | Get proposer of a proposal                                     | `GET /cosmos/gov/proposals/{proposal-id}` (Get proposer from `Proposal` struct)                       |
| `GET /gov/proposals/{proposal-id}/deposits`                            | Get deposits of a proposal                                     | `GET /cosmos/gov/proposals/{proposal-id}/deposits`                                                    |
| `GET /gov/proposals/{proposal-id}/deposits/{depositor}`                | Get depositor a of deposit                                     | `GET /cosmos/gov/proposals/{proposal-id}/deposits/{depositor}`                                        |
| `GET /gov/proposals/{proposal-id}/tally`                               | Get tally of a proposal                                        | `GET /cosmos/gov/proposals/{proposal-id}/tally`                                                       |
| `GET /gov/proposals/{proposal-id}/votes`                               | Get votes of a proposal                                        | `GET /cosmos/gov/proposals/{proposal-id}/votes`                                                       |
| `GET /gov/proposals/{proposal-id}/votes/{vote}`                        | Get a particular vote                                          | `GET /cosmos/gov/proposals/{proposal-id}/votes/{vote}`                                                |
| `GET /minting/parameters`                                              | Get parameters for minting                                     | `GET /cosmos/minting/params`                                                                          |
| `GET /minting/inflation`                                               | Get minting inflation                                          | `GET /cosmos/minting/inflation`                                                                       |
| `GET /minting/annual-provisions`                                       | Get minting annual provisions                                  | `GET /cosmos/minting/annual_provisions`                                                               |
| `POST /slashing/*`                                                     | Create unsigned `Msg`s for slashing                            | N/A, use Protobuf directly                                                                            |
| `GET /slashing/validators/{validatorPubKey}/signing_info`              | Get validator signing info                                     | `GET /cosmos/slashing/v1beta1/signing_infos/{cons_address}` (Use consensus address instead of pubkey) |
| `GET /slashing/signing_infos`                                          | Get all signing infos                                          | `GET /cosmos/slashing/v1beta1/signing_infos`                                                          |
| `GET /slashing/parameters`                                             | Get slashing parameters                                        | `GET /cosmos/slashing/v1beta1/params`                                                                 |
| `POST /staking/*`                                                      | Create unsigned `Msg`s for staking                             | N/A, use Protobuf directly                                                                            |
