// Package telemetry initializes OpenTelemetry and provides legacy metrics wrapper functions.
// End users only need to set the COSMOS_TELEMETRY environment variable to the path of
// an OpenTelemetry declarative configuration file: https://opentelemetry.io/docs/languages/sdk-configuration/declarative-configuration/
//
// Developers need to do two things:
//  1. Import this package before declaring any otel Tracer, Meter or Logger instances.
//  2. Make sure Shutdown() is called when the application is shutting down.
//     Tests can use the TestingInit function at startup to accomplish this.
//
// If these steps are followed, developers can follow the official golang otel conventions
// of declaring package-level tracer and meter instances using otel.Tracer() and otel.Meter().
// NOTE: it is important to thread context.Context properly for spans, metrics and logs to be
// correlated correctly.
// When using the SDK's context type, spans must be started with Context.StartSpan to
// get an SDK context which has the span set correctly.
// For logging, go.opentelemetry.io/contrib/bridges/otelslog provides a way to do this with the standard
// library slog package.
package telemetry
