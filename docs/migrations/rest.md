<!--
order: 2
-->

# REST Endpoints Migration

Migrate to gRPC-Gateway REST endpoints. Legacy REST endpoints were marked as deprecated in v0.40 and will be removed in v0.45. {synopsis}

::: warning
Two Legacy REST endpoints (`POST /txs` and `POST /txs/encode`) were removed ahead of schedule in v0.44 due to a security vulnerability.
:::

## Legacy REST Endpoints

Cosmos SDK versions v0.39 and earlier registered REST endpoints using a package called `gorilla/mux`. These REST endpoints were marked as deprecated in v0.40 and have since been referred to as legacy REST endpoints. Legacy REST endpoints will be officially removed in v0.45.

## gRPC-Gateway REST Endpoints

Following the Protocol Buffers migration in v0.40, Cosmos SDK has been set to take advantage of a vast number of gRPC tools and solutions. v0.40 introduced new REST endpoints generated from [gRPC `Query` services](../building-modules/query-services.md) using [grpc-gateway](https://grpc-ecosystem.github.io/grpc-gateway/). These new REST endpoints are referred to as gRPC-Gateway REST endpoints.

## Migrating to New REST Endpoints

| Legacy REST Endpoint                                                            | Description                                                         | New gRPC-gateway REST Endpoint                                                                        |
| ------------------------------------------------------------------------------- | ------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `GET /txs/{hash}`                                                               | Query tx by hash                                                    | `GET /cosmos/tx/v1beta1/txs/{hash}`                                                                   |
| `GET /txs`                                                                      | Query tx by events                                                  | `GET /cosmos/tx/v1beta1/txs`                                                                          |
| `POST /txs`                                                                     | Broadcast tx                                                        | `POST /cosmos/tx/v1beta1/txs`                                                                         |
| `POST /txs/encode`                                                              | Encodes an Amino JSON tx to an Amino binary tx                      | N/A, use Protobuf directly                                                                            |
| `POST /txs/decode`                                                              | Decodes an Amino binary tx into an Amino JSON tx                    | N/A, use Protobuf directly                                                                            |
| `POST /bank/*`                                                                  | Create unsigned `Msg`s for bank tx                                  | N/A, use Protobuf directly                                                                            |
| `GET /bank/balances/{address}`                                                  | Get the balance of an address                                       | `GET /cosmos/bank/v1beta1/balances/{address}`                                                 |
| `GET /bank/total`                                                               | Get the total supply of all coins                                   | `GET /cosmos/bank/v1beta1/supply`                                                                     |
| `GET /bank/total/{denom}`                                                       | Get the total supply of one coin                                    | `GET /cosmos/bank/v1beta1/supply/{denom}`                                                             |
| `POST /distribution/delegators/{delegatorAddr}/rewards`                         | Withdraw all delegator rewards                                      | N/A, use Protobuf directly                                                                            |
| `POST /distribution/*`                                                          | Create unsigned `Msg`s for distribution                             | N/A, use Protobuf directly                                                                            |
| `GET /distribution/delegators/{delegatorAddr}/rewards`                          | Get the total rewards balance from all delegations                  | `GET /cosmos/distribution/v1beta1/v1beta1/delegators/{delegator_address}/rewards`                     |
| `GET /distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}`          | Query a delegation reward                                           | `GET /cosmos/distribution/v1beta1/delegators/{delegatorAddr}/rewards/{validatorAddr}`                 |
| `GET /distribution/delegators/{delegatorAddr}/withdraw_address`                 | Get the rewards withdrawal address                                  | `GET /cosmos/distribution/v1beta1/delegators/{delegatorAddr}/withdraw_address`                        |
| `GET /distribution/validators/{validatorAddr}`                                  | Validator distribution information                                  | N/A                                         |
| `GET /distribution/validators/{validatorAddr}/rewards`                          | Commission and outstanding rewards of a single a validator      |  `GET /cosmos/distribution/v1beta1/validators/{validatorAddr}/commission` <br> `GET /cosmos/distribution/v1beta1/validators/{validatorAddr}/outstanding_rewards`                               |
| `GET /distribution/validators/{validatorAddr}/outstanding_rewards`              | Outstanding rewards of a single validator                           | `GET /cosmos/distribution/v1beta1/validators/{validatorAddr}/outstanding_rewards`                     |
| `GET /distribution/parameters`                                                  | Get the current distribution parameter values                       | `GET /cosmos/distribution/v1beta1/params`                                                             |
| `GET /distribution/community_pool`                                              | Get the amount held in the community pool                           | `GET /cosmos/distribution/v1beta1/community_pool`                                                     |
| `GET /evidence/{evidence-hash}`                                                 | Get evidence by hash                                                | `GET /cosmos/evidence/v1beta1/evidence/{evidence_hash}`                                               |
| `GET /evidence`                                                                 | Get all evidence                                                    | `GET /cosmos/evidence/v1beta1/evidence`                                                               |
| `POST /gov/*`                                                                   | Create unsigned `Msg`s for gov                                      | N/A, use Protobuf directly                                                                            |
| `GET /gov/parameters/{type}`                                                    | Get government parameters                                           | `GET /cosmos/gov/v1beta1/params/{type}`                                                               |
| `GET /gov/proposals`                                                            | Get all proposals                                                   | `GET /cosmos/gov/v1beta1/proposals`                                                                   |
| `GET /gov/proposals/{proposal-id}`                                              | Get proposal by id                                                  | `GET /cosmos/gov/v1beta1/proposals/{proposal-id}`                                                     |
| `GET /gov/proposals/{proposal-id}/proposer`                                     | Get proposer of a proposal                                          | N/A, use Query tx by events endpoint               |
| `GET /gov/proposals/{proposal-id}/deposits`                                     | Get deposits of a proposal                                          | `GET /cosmos/gov/v1beta1/proposals/{proposal-id}/deposits`                                            |
| `GET /gov/proposals/{proposal-id}/deposits/{depositor}`                         | Get depositor a of deposit                                          | `GET /cosmos/gov/v1beta1/proposals/{proposal-id}/deposits/{depositor}`                                |
| `GET /gov/proposals/{proposal-id}/tally`                                        | Get tally of a proposal                                             | `GET /cosmos/gov/v1beta1/proposals/{proposal-id}/tally`                                               |
| `GET /gov/proposals/{proposal-id}/votes`                                        | Get votes of a proposal                                             | `GET /cosmos/gov/v1beta1/proposals/{proposal-id}/votes`                                               |
| `GET /gov/proposals/{proposal-id}/votes/{vote}`                                 | Get a particular vote                                               | `GET /cosmos/gov/v1beta1/proposals/{proposal-id}/votes/{vote}`                                        |
| `GET /minting/parameters`                                                       | Get parameters for minting                                          | `GET /cosmos/minting/v1beta1/params`                                                                  |
| `GET /minting/inflation`                                                        | Get minting inflation                                               | `GET /cosmos/minting/v1beta1/inflation`                                                               |
| `GET /minting/annual-provisions`                                                | Get minting annual provisions                                       | `GET /cosmos/minting/v1beta1/annual_provisions`                                                       |
| `POST /slashing/*`                                                              | Create unsigned `Msg`s for slashing                                 | N/A, use Protobuf directly                                                                            |
| `GET /slashing/validators/{validatorPubKey}/signing_info`                       | Get validator signing info                                          | `GET /cosmos/slashing/v1beta1/signing_infos/{cons_address}` (Use consensus address instead of pubkey) |
| `GET /slashing/signing_infos`                                                   | Get all signing infos                                               | `GET /cosmos/slashing/v1beta1/signing_infos`                                                          |
| `GET /slashing/parameters`                                                      | Get slashing parameters                                             | `GET /cosmos/slashing/v1beta1/params`                                                                 |
| `POST /staking/*`                                                               | Create unsigned `Msg`s for staking                                  | N/A, use Protobuf directly                                                                            |
| `GET /staking/delegators/{delegatorAddr}/delegations`                           | Get all delegations from a delegator                                | `GET /cosmos/staking/v1beta1/delegations/{delegatorAddr}`                                  |
| `GET /staking/delegators/{delegatorAddr}/unbonding_delegations`                 | Get all unbonding delegations from a delegator                      | `GET /cosmos/staking/v1beta1/delegators/{delegatorAddr}/unbonding_delegations`                        |
| `GET /staking/delegators/{delegatorAddr}/txs`                                   | Get all staking txs (i.e msgs) from a delegator                     | Removed                                                                                               |
| `GET /staking/delegators/{delegatorAddr}/validators`                            | Query all validators that a delegator is bonded to                  | `GET /cosmos/staking/v1beta1/delegators/{delegatorAddr}/validators`                                   |
| `GET /staking/delegators/{delegatorAddr}/validators/{validatorAddr}`            | Query a validator that a delegator is bonded to                     | `GET /cosmos/staking/v1beta1/delegators/{delegatorAddr}/validators/{validatorAddr}`                   |
| `GET /staking/delegators/{delegatorAddr}/delegations/{validatorAddr}`           | Query a delegation between a delegator and a validator              | `GET /cosmos/staking/v1beta1/validators/{validatorAddr}/delegations/{delegatorAddr}`                  |
| `GET /staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}` | Query all unbonding delegations between a delegator and a validator | `GET /cosmos/staking/v1beta1/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}`        |
| `GET /staking/redelegations`                                                    | Query redelegations                                                 | `GET /cosmos/staking/v1beta1/v1beta1/delegators/{delegator_addr}/redelegations`                       |
| `GET /staking/validators`                                                       | Get all validators                                                  | `GET /cosmos/staking/v1beta1/validators`                                                              |
| `GET /staking/validators/{validatorAddr}`                                       | Get a single validator info                                         | `GET /cosmos/staking/v1beta1/validators/{validatorAddr}`                                              |
| `GET /staking/validators/{validatorAddr}/delegations`                           | Get all delegations to a validator                                  | `GET /cosmos/staking/v1beta1/validators/{validatorAddr}/delegations`                                  |
| `GET /staking/validators/{validatorAddr}/unbonding_delegations`                 | Get all unbonding delegations from a validator                      | `GET /cosmos/staking/v1beta1/validators/{validatorAddr}/unbonding_delegations`                        |
| `GET /staking/historical_info/{height}`                                         | Get HistoricalInfo at a given height                                | `GET /cosmos/staking/v1beta1/historical_info/{height}`                                                |
| `GET /staking/pool`                                                             | Get the current state of the staking pool                           | `GET /cosmos/staking/v1beta1/pool`                                                                    |
| `GET /staking/parameters`                                                       | Get the current staking parameter values                            | `GET /cosmos/staking/v1beta1/params`                                                                  |
| `POST /upgrade/*`                                                               | Create unsigned `Msg`s for upgrade                                  | N/A, use Protobuf directly                                                                            |
| `GET /upgrade/current`                                                          | Get the current plan                                                | `GET /cosmos/upgrade/v1beta1/current_plan`                                                            |
| `GET /upgrade/applied_plan/{name}`                                              | Get a previously applied plan                                       | `GET /cosmos/upgrade/v1beta1/applied/{name}`                                                          |

## Migrating to gRPC

Instead of hitting REST endpoints as described above, the Cosmos SDK also exposes a gRPC server. Any client can use gRPC instead of REST to interact with the node. An overview of different ways to communicate with a node can be found [here](../core/grpc_rest.md), and a concrete tutorial for setting up a gRPC client can be found [here](../run-node/txs.md#programmatically-with-go).