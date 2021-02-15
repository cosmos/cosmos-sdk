<!--
order: 7
-->

# Parameters

## Clients

The ibc clients contain the following parameters:

| Key              | Type | Default Value |
|------------------|------|---------------|
| `AllowedClients`    | []string | `"06-solomachine","07-tendermint"`        |

### AllowedClients

The allowed clients parameter defines an allowlist of client types supported by the chain. A client
that is not registered on this list will fail upon creation or on genesis validation. Note that,
since the client type is an arbitrary string, chains they must not register two light clients which
return the same value for the `ClientType()` function, otherwise the allowlist check can be
bypassed.
