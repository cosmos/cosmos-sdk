## Quick Start For Local Telemetry

To quickly set up a local telemetry environment where OpenTelemetry data is sent to a local instance of Grafana LGTM:

start the [Grafana LGTM docker image](https://hub.docker.com/r/grafana/otel-lgtm):

```shell
docker run -p 3000:3000 -p 4317:4317 -p 4318:4318 --rm -ti grafana/otel-lgtm
```

## Setting Up OpenTelemetry Configuration

We support two methods of setting up the OpenTelemetry configuration: via environment variable, and via node config.

## Example Configuration

Below is an example configuration for OpenTelemetry.

```yaml
file_format: "1.0-rc.3"
resource:
  attributes:
    - name: service.name
      value: simapp

tracer_provider:
  processors:
    - batch: # NOTE: you should use batch in production!
        exporter:
          otlp_grpc:
            endpoint: http://localhost:4317

meter_provider:
  readers:
    - pull:
        exporter:
          prometheus/development: # pushes directly to prometheus backend. 
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
          otlp_grpc:
            endpoint: http://localhost:4317


extensions:
  instruments:
    host: {} # enable optional host instrumentation with go.opentelemetry.io/contrib/instrumentation/host
    runtime: {} # enable optional runtime instrumentation with go.opentelemetry.io/contrib/instrumentation/runtime
    diskio: {} # enable optional disk I/O instrumentation using gopsutil
    # diskio with options:
    # diskio:
    #   disable_virtual_device_filter: true  # include virtual devices (loopback, RAID, partitions) on Linux
  propagators:
    - tracecontext
```

NOTE: the Go implementation may not support all options, so check the go [otelconf](https://pkg.go.dev/go.opentelemetry.io/contrib/otelconf) documentation carefully to see what is actually supported.

### Environment Variable

Using the environment variable method will instantiate the OpenTelemetry SDK before global meters and spans. 
This allows meters and traces to use direct references to the underlying instrument.

Set the `OTEL_EXPERIMENTAL_CONFIG_FILE` environment variable to the path of the configuration file:
   
Example: `export OTEL_EXPERIMENTAL_CONFIG_FILE=path/to/config.yaml`

## Node Config OpenTelemetry file

The node's `init` command will generate an empty otel file in the `~/.<node_home>/config` directory. 
You can paste your OpenTelemetry configuration here instead of using an environment variable. Note that when the environment variable is present, the configuration in the node config directory will be ignored.

## OpenTelemetry Initialization

While manual OpenTelemetry initialization is still supported, this package provides a single
point of initialization such that end users can just use the official
OpenTelemetry declarative configuration
spec: https://opentelemetry.io/docs/languages/sdk-configuration/declarative-configuration/
End users only need to set the `OTEL_EXPERIMENTAL_CONFIG_FILE` environment variable to the path of
an OpenTelemetry configuration file, or paste their config in the otel.yaml file in the node's config directory.
All the documentation for application instrumentation is provided in the OpenTelemetry documentation.

## Developer Usage

IMPORTANT: Make sure `telemetry.Shutdown()` is called when the application is shutting down.

If the steps above are followed, developers can follow the official Go OpenTelemetry conventions
of declaring package-level tracer and meter instances using `otel.Tracer()` and `otel.Meter()`.
NOTE: it is important to thread `context.Context` properly for spans and metrics to be correlated.
When using the SDK's context type, spans must be started with the context's `StartSpan` method to
get an SDK context which has the span set correctly.

```go
import (
	"go.opentelemetry.io/otel"

    sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
    tracer = otel.Tracer("cosmos-sdk/baseapp")
)

func Example(ctx sdk.Context) {
   ctx, span := ctx.StartSpan(tracer, "VerifyVoteExtension")
   defer span.End()
   // ....
}
```