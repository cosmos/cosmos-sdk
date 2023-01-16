import { FieldType, FormField } from 'src/components/inputs/Inputs';

import { RunLoadtestRequest } from 'src/gen/orijtech/cosmosloadtester/v1/loadtest_service_pb';

const BroadcastTxMethod = RunLoadtestRequest.BroadcastTxMethod;
const SelectMethod = RunLoadtestRequest.EndpointSelectMethod;

export const fields: FormField[] = [
    {
        name: 'broadcastTxMethod',
        label: 'Broadcast TX Method',
        info: 'The broadcast tx method to use when submitting transactions, can be async, sync, or commit',
        fieldType: FieldType.SINGLE_SELECTION_BASED,
        default: BroadcastTxMethod.BROADCAST_TX_METHOD_ASYNC,
        options: [
            { name: 'Async', value: BroadcastTxMethod.BROADCAST_TX_METHOD_ASYNC },
            { name: 'Sync', value: BroadcastTxMethod.BROADCAST_TX_METHOD_SYNC },
            { name: 'Commit', value: BroadcastTxMethod.BROADCAST_TX_METHOD_COMMIT },
        ],
    },
    {
        name: 'endpointSelectMethod',
        label: 'Endpoint Select method',
        info: 'The method by which to select endpoints. Maps to --endpoint-select-method in tm-load-test.',
        fieldType: FieldType.SINGLE_SELECTION_BASED,
        default: SelectMethod.ENDPOINT_SELECT_METHOD_SUPPLIED,
        options: [
            { name: 'Supplied', value: SelectMethod.ENDPOINT_SELECT_METHOD_SUPPLIED, info: 'select only supplied endpoints(s) for load testing' },
            { name: 'Discovered', value: SelectMethod.ENDPOINT_SELECT_METHOD_DISCOVERED, info: 'select newly discovered endpoints only (excluding supplied endpoints)' },
            { name: 'Any', value: SelectMethod.ENDPOINT_SELECT_METHOD_ANY, info: 'select from any of supplied and/or discovered endpoints' }
        ]
    },
    {
        name: 'clientFactory',
        required: 'Client factory field is required.',
        label: 'Client factory',
        default: 'test-cosmos-client-factory',
        info: 'The identifier of the client factory to use for generating load testing transactions. Maps to --client-factory in tm-load-test',
        fieldType: FieldType.VALUE_BASED,
    },
    {
        name: 'duration',
        label: 'Duration',
        info: 'The duration (in seconds) for which to handle the load test. Maps to --time in tm-load-test.',
        required: 'Duration to handle load test is required',
        default: 20,
        fieldType: FieldType.TIME_BASED,
    },
    {
        name: 'endpoints',
        label: 'Endpoints',
        required: 'Endpoint(s) required to run load test.',
        info: 'A comma-separated list of URLs indicating Tendermint WebSockets RPC endpoints to which to connect. Maps to --endpoints in tm-load-test.',
        fieldType: FieldType.LIST_BASED,
    },
    {
        name: 'maxEndpointCount',
        label: 'Max endpoint count',
        info: 'The maximum number of endpoints to use for testing, where 0 means unlimited. Maps to --max-endpoints in tm-load-test.',
        default: 10,
        fieldType: FieldType.VALUE_BASED,
    },
    {
        name: 'minPeerConnectivityCount',
        label: 'Min peer connectivity count',
        info: 'The minimum number of peers to which each peer must be connected before starting the load test. Maps to --min-peer-connectvity in tm-load-test.',
        default: 0,
        fieldType: FieldType.VALUE_BASED,
    },
    {
        name: 'peerConnectTimeout',
        label: 'Peer connect timeout',
        info: 'The number of seconds to wait for all required peers to connect if expect-peers > 0. Maps to --peer-connect-timeout in tm-load-test.',
        default: 20,
        fieldType: FieldType.TIME_BASED,
    },
    {
        name: 'sendPeriod',
        label: 'Send period',
        info: 'The period (in seconds) at which to send batches of transactions. Maps to --send-period in tm-load-test.',
        default: 1,
        fieldType: FieldType.TIME_BASED,
    },
    {
        name: 'connectionCount',
        label: 'Connection Count',
        info: 'The number of connections to open to each endpoint simultaneously. Maps to --connections in tm-load-test.',
        default: 1,
        fieldType: FieldType.VALUE_BASED,
    },
    {
        name: 'statsOutputFilePath',
        label: 'Stats Output File Path',
        info: 'Where to store aggregate statistics (in CSV format) for the load test. Maps to --stats-output in tm-load-test.',
        default: "",
        fieldType: FieldType.VALUE_BASED,
    },
    {
        name: 'expectPeersCount',
        label: 'Expect peers count',
        info: 'The minimum number of peers to expect when crawling the P2P network from the specified endpoint(s) prior to waiting for workers to connect. Maps to --expect-peers in tm-load-test.',
        default: 0,
        fieldType: FieldType.VALUE_BASED,
    },
    {
        name: 'transactionSizeBytes',
        label: 'Transaction size bytes',
        info: 'The size of each transaction, in bytes - must be greater than 40. Maps to --size in tm-load-test.',
        default: 250,
        fieldType: FieldType.VALUE_BASED,
    },
    {
        name: 'transactionsPerSecond',
        label: 'Transaction rate per second',
        info: 'The number of transactions to generate each second on each connection, to each endpoint. Maps to --rate in tm-load-test.',
        default: 1000,
        fieldType: FieldType.VALUE_BASED,
    },
    {
        name: 'transactionCount',
        label: 'Transaction count',
        info: 'The maximum number of transactions to send - set to -1 to turn off this limit. Maps to --count in tm-load-test.',
        default: -1,
        fieldType: FieldType.VALUE_BASED,
    },
];
