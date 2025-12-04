## Quick Start For Local Telemetry

To quickly setup a local telemetry environment where OpenTelemetry data is sent to a local instance of Grafana LGTM:

start the [Grafana LGTM docker image](https://hub.docker.com/r/grafana/otel-lgtm):

```shell
docker run -p 3000:3000 -p 4317:4317 -p 4318:4318 --rm -ti grafana/otel-lgtm
```
## Environment Variable

Using the environment variable method will instantiate the OpenTelemetry SDK before global meters and spans. 
This allows meters and traces to use direct references to the underlying instrument.

Create a basic OpenTelemetry configuration file which will send data to the local instance of Grafana LGTM:

```yaml
resource:
  attributes:
    - name: service.name
      value: simapp

tracer_provider:
  processors:
    - batch: # NOTE: you should use batch in production!
        exporter:
          otlp:
            protocol: grpc
            endpoint: http://localhost:4317

meter_provider:
  readers:
    - pull:
        exporter:
          prometheus: # pushes directly to prometheus backend. 
            host: 0.0.0.0
            port: 9464
            # optional: include resource attributes as constant labels
            with_resource_constant_labels:
              include:
                - service.name
  views:
    - selector:
        instrument_type: histogram
      stream:
        aggregation:
          explicit_bucket_histogram:
            boundaries: [ 0.2, 0.25, 0.3, 0.35, 0.4, 0.5, 0.75, 1, 2, 5 ]

logger_provider:
  processors:
    - batch:
        exporter:
          otlp:
            protocol: grpc
            endpoint: http://localhost:4317


cosmos_extra:
  trace_file: ""
  metrics_file: ""
  metrics_file_interval: ""
  logs_file: ""
  instrument_host: true
  instrument_runtime: true
  propagators:
    - tracecontext
```

For a full list of configurable options see: https://github.com/open-telemetry/opentelemetry-configuration/blob/main/examples/kitchen-sink.yaml

3. set the `OTEL_EXPERIMENTAL_CONFIG_FILE` environment variable to the path of the configuration file:
   `export OTEL_EXPERIMENTAL_CONFIG_FILE=path/to/config.yaml`
4. start your application or tests
5. view the data in Grafana LGTM at http://localhost:3000/. The Drilldown views are suggested for getting started.

## Node Home OpenTelemetry file

The node's `init` command will generate an empty otel file in the `~/.<node_home>/config` directory. Place your otel configuration
here. 

When the node's `start` command is ran, the OpenTelemetry SDK will be initialized using this file. 
If left empty, all meters and tracers will be noop.

## OpenTelemetry Initialization

While manual OpenTelemetry initialization is still supported, this package provides a single
point of initialization such that end users can just use the official
OpenTelemetry declarative configuration
spec: https://opentelemetry.io/docs/languages/sdk-configuration/declarative-configuration/
End users only need to set the `OTEL_EXPERIMENTAL_CONFIG_FILE` environment variable to the path of
an OpenTelemetry configuration file, or fill out the otel.yaml file in the node's config directory and that's it.
All the documentation necessary is provided in the OpenTelemetry documentation.

## Developer Usage

If using the environment variable method, importing the baseapp package will cause the telemetry's initialization to run.
Otherwise, ensure the otel.yaml file in the node's config directory is filled out.

IMPORTANT: Make sure Shutdown() is called when the application is shutting down.

Tests can use the TestingInit function at startup to accomplish this.

If these steps are followed, developers can follow the official golang otel conventions
of declaring package-level tracer and meter instances using otel.Tracer() and otel.Meter().
NOTE: it is important to thread context.Context properly for spans and metrics to be
correlated correctly.
When using the SDK's context type, spans must be started with Context.StartSpan to
get an SDK context which has the span set correctly.

