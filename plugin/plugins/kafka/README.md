# Kafka Indexing Plugin

This plugin demonstrates how to listen to state changes of individual `KVStores` as described in [ADR-038 State Listening](https://github.com/vulcanize/cosmos-sdk/blob/adr038_plugin_proposal/docs/architecture/adr-038-state-listening.md) and index the data in Kafka.



<!-- TOC -->
  - [Plugin design](#plugin-design)
    - [Channel-Based producer](#channel-based-producer)
    - [Delivery Report handler](#delivery-report-handler)
    - [Message serde](#message-serde)
    - [Example configuration](#example-configuration)
  - [Testing the plugin](#testing-the-plugin)
  - [Confluent Platform](#confluent-platform)
    - [Docker](#docker)
    - [Schema Registry](#schema-registry)
    - [KSQL examples](#ksql-examples)
    


## Plugin design
The plugin was build using [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go); a lightwieght wrapper around [librdkafka](https://github.com/edenhill/librdkafka).

This particular implmentation uses:
* `Channel-Based producer` - Faster than the function-based `producer.Produce()`.
* `Delivery reports handler` - Notifies the application of success or failure to deliver messages to Kafka.

### Channel-Based producer
The plugin uses the `producer.Producerchannel()` to deliver messages to Kafka.


Pros:
* Proper channel backpressure if `librdkafka`'s internal queue is full. The queue size can be controlled by setting.
* Message order is preserved (guaranteed by the producer API).
* Faster than the `function-based` async producer.

Cons:
* Double queueing: messages are first queued in the channel and the inside librdkafka. the Size of the channel is configurable via `queue.buffering.max.messages`.

### Delivery Report handler
Producing is an asynchronous operation. Therefore, the client notifies the application (per-message) of success or failure through delivery reports. Deliver reports are by default emmitted on the `producer.Events()` channel as `*kafka.Message`. One needs to check `msg.TopicPartition.Error` for `nil` to find out if the message was successfully delivered or not.

Pros:
* Can be used to propagate success or failures to the application.
* Can be used to track the messages produced.
* Can be turned off by setting `"go.delivery.reports": false` for a fire-and-forget scenario.

Cons:
* Must be handled in a go routine which makes it difficult to propagate errors to the `WriterListner.onWrite()`.

### Message serde

As of this writing there is no `golang` support for `serialization/deserialization` of proto messages for the Confluent Schema Registry. Because of this limitiation, the Marshalled JSON data is saved instead.

Note, you can register the proto messages with the schema registry by generating the `Java` code and using the supported [Java](https://github.com/confluentinc/schema-registry/blob/master/protobuf-serializer/src/main/java/io/confluent/kafka/serializers/protobuf/KafkaProtobufSerializer.java) client library for the schema registry to automatically register the proto messages.

#### Message `key`
To be able to identify an track messages in Kafka the `key` is made up of the following properties:
* `block height` - BIGINT
* `event` - BEGIN_BLOCK, END_BLOCK, DELIVER_TX
* `event_id` - BIGINT (increments for DELIVER_TX)
* `event_type` - REQUEST, RESPONSE, STATE_CHANGE
* `event_type_id` - BIGINT (increments for STATE_CHANGE)

Example:
```
// first tx
{
  "block_height": 1,
  "event": "DELIVER_TX",
  "event_id": 1,
  "event_type": "REQUEST",
  "event_type_id ": 1
}

// second tx
{
  "block_height": 1,
  "event": "DELIVER_TX",
  "event_id": 2,           // incrementing
  "event_type": "REQUEST",
  "event_type_id ": 1
}
```

#### Message `value`

The `value` structure is the Marshalled JSON structure of the request, response or the state change for begin block, end block, and deliver tx events.

Example:
```
{
  "BLOCK_HEIGHT": 1,
  "EVENT": "BEGIN_BLOCK",
  "EVENT_ID": 1,
  "EVENT_TYPE": "STATE_CHANGE",
  "EVENT_TYPE_ID": 1,
  "STORE_KEY": "mockStore1",
  "DELETE": false,
  "KEY": "AQID",
  "VALUE": "AwIB"
}
```

### Example configuration

Below is an example of how to configure the Kafka plugin.
```
# app.toml

...

###############################################################################
###                      Plugin system configuration                        ###
###############################################################################

[plugins]

# turn the plugin system, as a whole, on or off
on = true

# List of plugin names to enable from the plugin/plugins/*
enabled = ["kafka"]

# The directory to load non-preloaded plugins from; defaults to ./plugin/plugins
dir = ""

# a mapping of plugin-specific streaming service parameters, mapped to their pluginFileName
[plugins.streaming]

# maximum amount of time the BaseApp will await positive acknowledgement of message receipt from all streaming services
# in milliseconds
global_ack_wait_limit = 2000

###############################################################################
###                       Kafka Plugin configuration                        ###
###############################################################################

# The specific parameters for the Kafka streaming service plugin
[plugins.streaming.kafka]

# List of store keys we want to expose for this streaming service.
keys = []

# Optional topic prefix for the topic(s) where data will be stored
topic_prefix = "block"

# Flush and wait for outstanding messages and requests to complete delivery. (milliseconds)
flush_timeout_ms = 1500

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
client_id = "my-app-id"

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

## Confluent Platform

If you're interested in viewing or querying events stored in kafka you can stand up the Confluent Platform stack with docker.

*Visit the Confluent Platform [docs](https://docs.confluent.io/platform/current/quickstart/ce-docker-quickstart.html) for up to date docker instructions.*

### Docker

Spin up Confluent Platform.
```
docker-compose -f plugin/plugins/kafka/docker-compose.yml up -d
```

You should see something like this:
```
Creating network "kafka_default" with the default driver
Creating zookeeper ... done
Creating broker    ... done
Creating schema-registry ... done
Creating rest-proxy      ... done
Creating connect         ... done
Creating ksqldb-server   ... done
Creating ksql-datagen    ... done
Creating ksqldb-cli      ... done
Creating control-center  ... done
```

Check status
```
docker-compose ps
```

You should see something like this:
```
     Name                    Command               State                       Ports                     
---------------------------------------------------------------------------------------------------------
broker            /etc/confluent/docker/run        Up      0.0.0.0:9092->9092/tcp, 0.0.0.0:9101->9101/tcp
connect           /etc/confluent/docker/run        Up      0.0.0.0:8083->8083/tcp, 9092/tcp              
control-center    /etc/confluent/docker/run        Up      0.0.0.0:9021->9021/tcp                        
ksql-datagen      bash -c echo Waiting for K ...   Up                                                    
ksqldb-cli        /bin/sh                          Up                                                    
ksqldb-server     /etc/confluent/docker/run        Up      0.0.0.0:8088->8088/tcp                        
rest-proxy        /etc/confluent/docker/run        Up      0.0.0.0:8082->8082/tcp                        
schema-registry   /etc/confluent/docker/run        Up      0.0.0.0:8081->8081/tcp                        
zookeeper         /etc/confluent/docker/run        Up      0.0.0.0:2181->2181/tcp, 2888/tcp, 3888/tcp  
```

### KSQL examples

One huge advante of using Kafka with the Confluent Platform is the KSQL streaming engine. KSQL allows us to be able to write queries and create streams or tables from one or multiple Kafka topics (through joins) without having to write any code.

Examples:

Create a structured stream from the `block-state-change` topic containig raw data. This will make it easier to be able to fitler out specific events.
```
CREATE OR REPLACE STREAM state_change_stream (
  block_height  BIGINT KEY,    /* k1 */
  event         STRING KEY,    /* k2 */
  event_id      BIGINT KEY,    /* k3 */
  event_type    STRING KEY,    /* k4 */
  event_type_id BIGINT KEY,    /* k5 */
  store_key     STRING,
  `delete`      BOOLEAN,
  key           STRING,
  value         STRING         /* this may be a STRUC depending on the store type */
) WITH (KAFKA_TOPIC='block-state-change', KEY_FORMAT='JSON', VALUE_FORMAT='JSON');
```

Run the below query to see the messages in of this new stream.

```
SELECT * FROM state_change_stream EMIT CHANGES LIMIT 1;
```

Result:
```
{
  "BLOCK_HEIGHT": 1,
  "EVENT": "BEGIN_BLOCK",
  "EVENT_ID": 1,
  "EVENT_TYPE": "STATE_CHANGE",
  "EVENT_TYPE_ID": 1,
  "STORE_KEY": "mockStore1",
  "delete": false,
  "KEY": "AQID",
  "VALUE": "AwIB"
}
```

Lets take it one step further and create a stream that contains only `DELIVER_TX` events.

```
SET 'processing.guarantee' = 'exactly_once';

CREATE OR REPLACE STREAM deliver_tx_state_change_stream
  AS SELECT * 
  FROM  STATE_CHANGE_STREAM 
  WHERE event = 'DELIVER_TX'
  EMIT CHANGES;
```

Lets take a look at what the data looks like.

```
SELECT * FROM deliver_tx_state_change_stream EMIT CHANGES LIMIT 1;
```

Result:

```
{
  "BLOCK_HEIGHT": 2,
  "EVENT": "BEGIN_BLOCK",
  "EVENT_ID": 1,
  "EVENT_TYPE": "STATE_CHANGE",
  "EVENT_TYPE_ID": 1,
  "STORE_KEY": "acc",
  "delete": false,
  "KEY": "AQBhNv4khMI7PylvV6i1lSlSCleL",
  "VALUE": "CiAvY29zbW9zLmF1dGgudjFiZXRhMS5CYXNlQWNjb3VudBJ8Ci1jb3Ntb3MxcXBzbmRsM3lzbnByazBlZmRhdDYzZHY0OTlmcTU0dXR0eWdncGsSRgofL2Nvc21vcy5jcnlwdG8uc2VjcDI1NmsxLlB1YktleRIjCiECcyIkZHE6G+gkK2TJEjko3LjNFgZ4Fmfu90jDkjlbojcYygEgAQ=="
}
```

### Schema Registry

Because `golang` lacks support to be able to register Protobuf messages with the schema registry, one needs to generate the Java code from the proto messages and use the [KafkaProtobufSerializer.java](https://github.com/confluentinc/schema-registry/blob/master/protobuf-serializer/src/main/java/io/confluent/kafka/serializers/protobuf/KafkaProtobufSerializer.java) to automatically register them. The Java libraries make this process exctreamly easy. Take a look [here](https://docs.confluent.io/platform/current/schema-registry/serdes-develop/serdes-protobuf.html) fro an example of how this is achived.


Check out the [docs](https://docs.ksqldb.io/en/latest/) and this [post](https://www.confluent.io/blog/ksqldb-0-15-reads-more-message-keys-supports-more-data-types/) for more complex examples and a deeper understanding of KSQL.
