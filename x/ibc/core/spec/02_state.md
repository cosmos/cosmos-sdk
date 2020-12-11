<!--
order: 2
-->

# State

The paths for the values stored in state is defined [here](https://github.com/cosmos/ics/blob/master/spec/ics-024-host-requirements/README.md#path-space).
Additionally, the SDK adds a prefix to the path to be able to aggregate the values for querying purposes.
The client type is not stored since it can be obtained through the client state. 

| Prefix | Path                                                                        | Value type     |
|--------|-----------------------------------------------------------------------------|----------------|
| "0/"   | "clients/{identifier}/clientState"                                          | ClientState    |
| "0/"   | "clients/{identifier}/consensusStates/{height}"                             | ConsensusState |
| "0/"   | "clients/{identifier}/connections"                                          | []string       |
| "0/"   | "nextClientSequence                                                         | uint64         |
| "0/"   | "connections/{identifier}"                                                  | ConnectionEnd  |
| "0/"   | "nextConnectionSequence"                                                    | uint64         |
| "0/"   | "ports/{identifier}"                                                        | CapabilityKey  |
| "0/"   | "channelEnds/ports/{identifier}/channels/{identifier}"                      | ChannelEnd     |
| "0/"   | "nextChannelSequence"                                                       | uint64         |
| "0/"   | "capabilities/ports/{identifier}/channels/{identifier}"                     | CapabilityKey  |
| "0/"   | "nextSequenceSend/ports/{identifier}/channels/{identifier}"                 | uint64         |
| "0/"   | "nextSequenceRecv/ports/{identifier}/channels/{identifier}"                 | uint64         |
| "0/"   | "nextSequenceAck/ports/{identifier}/channels/{identifier}"                  | uint64         |
| "0/"   | "commitments/ports/{identifier}/channels/{identifier}/sequences/{sequence}" | bytes          |
| "0/"   | "receipts/ports/{identifier}/channels/{identifier}/sequences/{sequence}"    | bytes          |
| "0/"   | "acks/ports/{identifier}/channels/{identifier}/sequences/{sequence}"        | bytes          |
