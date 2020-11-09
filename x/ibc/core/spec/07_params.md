<!--
order: 7
-->

# Parameters

## Clients

The ibc clients contain the following parameters:

| Key              | Type | Default Value |
|------------------|------|---------------|
| `AllowedClients`    | []string | `"Solo Machine","Tendermint"`        |

### AllowedClients

The allowed clients parameter defines an allowlist of client types supported by the chain. A client
that is not registered on this list will fail upon creation or on genesis validation.
