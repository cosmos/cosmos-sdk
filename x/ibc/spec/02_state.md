<!--
order: 2
-->

# State

The paths for the values stored in state can be found [here](https://github.com/cosmos/ics/blob/master/spec/ics-024-host-requirements/README.md#path-space). Additionally, the SDK adds
a prefix to the path to be able to aggregate the values for querying purposes.

| Prefix | Path                                                                   | Value type     |
|--------|------------------------------------------------------------------------|----------------|
| "0/"   | "clients/{identifier}"                                                 | ClientState    |
| "0/"   | "clients/{identifier}/consensusState"                                  | ConsensusState |
| "0/"   | "clients/{identifier}/type"                                            | ClientType     |
| "0/"   | "connections/{identifier}"                                             | ConnectionEnd  |
| "0/"   | "ports/{identifier}"                                                   | CapabilityKey  |
| "0/"   | "ports/{identifier}/channels/{identifier}"                             | ChannelEnd     |
| "0/"   | "ports/{identifier}/channels/{identifier}/key"                         | CapabilityKey  |
| "0/"   | "ports/{identifier}/channels/{identifier}/nextSequenceRecv"            | uint64         |
| "0/"   | "ports/{identifier}/channels/{identifier}/packets/{sequence}"          | bytes          |
| "0/"   | "ports/{identifier}/channels/{identifier}/acknowledgements/{sequence}" | bytes          |