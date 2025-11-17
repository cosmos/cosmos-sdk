## Quick Start For Local Telemetry

To quickly setup a local telemetry environment where OpenTelemetry data is sent to a local instance of Grafana LGTM:
1. start the [Grafana LGTM docker image](https://hub.docker.com/r/grafana/otel-lgtm):
```shell
docker run -p 3000:3000 -p 4317:4317 -p 4318:4318 --rm -ti grafana/otel-lgtm
```
2. create a basic OpenTelemetry configuration file which will send data to the local instance of Grafana LGTM:
```yaml
resource:
  attributes:
    - name: service.name
      value: my_app_name
tracer_provider:
  processors:
    - batch: # NOTE: you should use batch in production!
        exporter:
          otlp:
            protocol: grpc
            endpoint: http://localhost:4317
meter_provider:
  readers:
    - periodic:
        interval: 1000 # 1 second, maybe use something longer in production
        exporter:
          otlp:
            protocol: grpc
            endpoint: http://localhost:4317
logger_provider:
  processors:
    - batch:
        exporter:
          otlp:
            protocol: grpc
            endpoint: http://localhost:4317
```
3. set the `OTEL_EXPERIMENTAL_CONFIG_FILE` environment variable to the path of the configuration file:
`export OTEL_EXPERIMENTAL_CONFIG_FILE=path/to/config.yaml`
4. start your application or tests
5. view the data in Grafana LGTM at http://localhost:3000/. The Drilldown views are suggested for getting started.

## OpenTelemetry Initialization

While manual OpenTelemetry initialization is still supported, this package provides a single
point of initialization such that end users can just use the official
OpenTelemetry declarative configuration spec: https://opentelemetry.io/docs/languages/sdk-configuration/declarative-configuration/
End users only need to set the `OTEL_EXPERIMENTAL_CONFIG_FILE` environment variable to the path of
an OpenTelemetry configuration file and that's it.
All the documentation necessary is provided in the OpenTelemetry documentation.

## Developer Usage

Developers need to do two things to use this package properly:
 1. Import this package before declaring any otel Tracer, Meter or Logger instances.
 2. Make sure Shutdown() is called when the application is shutting down.
    Tests can use the TestingInit function at startup to accomplish this.

If these steps are followed, developers can follow the official golang otel conventions
of declaring package-level tracer and meter instances using otel.Tracer() and otel.Meter().
NOTE: it is important to thread context.Context properly for spans, metrics, and logs to be
correlated correctly.
When using the SDK's context type, spans must be started with Context.StartSpan to
get an SDK context which has the span set correctly.
For logging, go.opentelemetry.io/contrib/bridges/otelslog provides a way to do this with the standard
library slog package.
