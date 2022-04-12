# Kafka Indexing Plugin

This plugin demonstrates how to listen to state changes of individual `KVStores` as described in [ADR-038 State Listening](https://github.com/vulcanize/cosmos-sdk/blob/adr038_plugin_proposal/docs/architecture/adr-038-state-listening.md) and index the data in Kafka.



<!-- TOC -->
  - [Plugin design](#plugin-design)
    - [Function-Based producer](#function-based-producer)
    - [Delivery Report handler](#delivery-report-handler)
    - [Message serde](#message-serde)
    - [Message key](#message-key)
    - [Example configuration](#example-configuration)
  - [Testing the plugin](#testing-the-plugin)
    - [Confluent Platform](#confluent-platform)
    


## Plugin design
The plugin was build using [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go), a lightweight wrapper around [librdkafka](https://github.com/edenhill/librdkafka).

This particular implementation uses:
* `Channel-Based producer` - Faster than the function-based `producer.Produce()`.
* `Delivery reports handler` - Notifies the application of success or failure to deliver messages to Kafka.

### Function-Based producer
The plugin uses the `producer.Produce()` to deliver messages to Kafka. Delivery reports are emitted on the `producer.Events()` or specific private channel.
Any errors that occur during delivery propagate up the stack and `halt` the app when `plugins.streaming.kafka.halt_app_on_delivery_error = true`

Pros:
* Go:ish

Cons:
* `Produce()` is a non-blocking call, if the internal librdkafka queue is full the call will fail.

*The Producer's queue is configurable with the `queue.buffering.max.messages` property (default: 100000). See [config-docs](https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md) for further understanding.

### Delivery Report handler
Producing is an asynchronous operation. Therefore, the client notifies the application (per-message) of success or failure through delivery reports.
Deliver reports are by default emitted on the `producer.Events()` channel as `*kafka.Message`. One needs to check `msg.TopicPartition.Error` for `nil` for successful delivery.
The plugin implementation uses a private delivery channel `Produce(msg *Message, deliveryChan chan Event)` for successful delivery of each message.
When `plugins.streaming.kafka.halt_app_on_delivery_error = true`, the app will `halt` if delivery of any messages fails. This helps keep state in sync between the node and Kafka.

Pros:
* Used to propagate success or failures to the application.
* Used to track the messages produced.
* Is turned off by setting `"go.delivery.reports": false` for a fire-and-forget scenario.

Cons:
* Slower than when the plugin operates in fire-and-forget mode `plugins.streaming.kafka.halt_app_on_delivery_error = false` as each message needs to be checked whether it was successfully delivered. 

### Message serde

As of this writing there is no `golang` support for `serialization/deserialization` of proto message schema with the Confluent Schema Registry. Therefore, the Kafka plugin produces messages in proto binary format without a registered schema.

Note, proto message schemas can be registered with the Confluent Schema Registry by [generating the Java code](https://developers.google.com/protocol-buffers/docs/javatutorial) of the CosmosSDK proto files and then use the supported Java libraries. See the Confluent [docs](https://docs.confluent.io/platform/current/schema-registry/serdes-develop/serdes-protobuf.html) for how to do this.

#### Message `key`

To be able to identify and track messages in Kafka a [msg_key.proto](./proto/msg_key.proto) was introduced to the plugin.

```
syntax = "proto3";

option go_package = "github.com/cosmos/cosmos-sdk/plugin/plugins/kafka/service";

option java_multiple_files = true;
option java_package = "network.cosmos.listening.plugins.kafka.service";

message MsgKey {
  int64  block_height  = 1 [json_name = "block_height"];
  enum Event {
    BEGIN_BLOCK = 0;
    END_BLOCK   = 1;
    DELIVER_TX  = 2;
  }
  Event event = 2;
  int64  event_id      = 3 [json_name = "event_id"];
  enum EventType {
    REQUEST      = 0;
    RESPONSE     = 1;
    STATE_CHANGE = 2;
  }
  EventType event_type = 4 [json_name = "event_type"];
  int64  event_type_id = 5 [json_name = "event_type_id"];
}
```

### Example configuration

Below is an example of how to configure the Kafka plugin.
```
# app.toml

. . .

###############################################################################
###                      Plugin system configuration                        ###
###############################################################################

[plugins]

# turn the plugin system, as a whole, on or off
on = true

# List of plugin names to enable from the plugin/plugins/*
enabled = ["kafka"]

# The directory to load non-preloaded plugins from; defaults $GOPATH/src/github.com/cosmos/cosmos-sdk/plugin/plugins
dir = ""


###############################################################################
###                       Kafka Plugin configuration                        ###
###############################################################################

# The specific parameters for the kafka streaming service plugin
[plugins.streaming.kafka]

# List of store keys we want to expose for this streaming service.
keys = []

# Optional prefix for topic names where data will be stored.
topic_prefix = "pio"

# Flush and wait for outstanding messages and requests to complete delivery. (milliseconds)
flush_timeout_ms = 5000

# Whether or not to halt the application when plugin fails to deliver message(s).
halt_app_on_delivery_error = true

# Producer configuration properties.
# The plugin uses confluent-kafka-go which is a lightweight wrapper around librdkafka.
# For a full list of producer configuration properties
# see https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
[plugins.streaming.kafka.producer]

# Initial list of brokers as a comma seperated list of broker host or host:port[, host:port[,...]]
bootstrap_servers = "localhost:9092"

# Client identifier
client_id = "pio-state-listening"

# This field indicates the number of acknowledgements the leader
# broker must receive from ISR brokers before responding to the request
acks = "all"

# When set to true, the producer will ensure that messages
# are successfully produced exactly once and in the original produce order.
# The following configuration properties are adjusted automatically (if not modified by the user)
# when idempotence is enabled: max.in.flight.requests.per.connection=5 (must be less than or equal to 5),
# retries=INT32_MAX (must be greater than 0), acks=all, queuing.strategy=fifo.
# Producer instantation will fail if user-supplied configuration is incompatible.
enable_idempotence = true
```

## Testing the plugin

Non-determinism testing has been set up to run with the Kafka plugin.

To execute the tests, run:
```
make test-sim-nondeterminism-state-listening-kafka
```

### Confluent Platform

If you're interested in viewing or querying events stored in kafka you can stand up the Confluent Platform stack with docker.
Visit the Confluent Platform [docs](https://docs.confluent.io/platform/current/quickstart/ce-docker-quickstart.html) for up to date docker instructions.
