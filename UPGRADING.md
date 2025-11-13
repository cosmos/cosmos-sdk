# Upgrade Reference

This document provides a quick reference for the upgrades from `v0.53.x` to `v0.54.x` of Cosmos SDK.

Note, always read the **App Wiring Changes** section for more information on application wiring updates.

🚨Upgrading to v0.54.x will require a **coordinated** chain upgrade.🚨

### TLDR

**The only major feature in Cosmos SDK v0.54.x is the upgrade from CometBFT v0.x.x to CometBFT v2.**

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.54.x/CHANGELOG.md).

#### Deprecation of `TimeoutCommit`

CometBFT v2 has deprecated the use of `TimeoutCommit` for a new field, `NextBlockDelay`, that is part of the
`FinalizeBlockResponse` ABCI message that is returned to CometBFT via the SDK baseapp.  More information from
the CometBFT repo can be found [here](https://github.com/cometbft/cometbft/blob/88ef3d267de491db98a654be0af6d791e8724ed0/spec/abci/abci%2B%2B_methods.md?plain=1#L689).

For SDK application developers and node runners, this means that the `timeout_commit` value in the `config.toml` file
is still used if `NextBlockDelay` is 0 (its default value).  This means that when upgrading to Cosmos SDK v0.54.x, if 
the existing `timout_commit` values that validators have been using will be maintained and have the same behavior.

For setting the field in your application, there is a new `baseapp` option, `SetNextBlockDelay` which can be passed to your application upon
initialization in `app.go`.  Setting this value to any non-zero value will override anything that is set in validators' `config.toml`.

#### Adoption of OpenTelemetry and Deprecation of `github.com/hashicorp/go-metrics`

Existing Cosmos SDK telemetry support is provide by `github.com/hashicorp/go-metrics` which is undermaintained and only supported metrics instrumentation.
OpenTelemetry provides an integrated solution for metrics, traces and logging which is widely adopted and actively maintained.
Also the existing wrapper functions in the `telemetry` package required acquiring mutex locks and map lookups for every metric operation which is sub-optimal. OpenTelemetry's API uses atomic concurrency wherever possible and should introduce less performance overhead during metric collection.

The [README.md](telemetry/README.md) in the `telemetry` package provides more details on usage, but below is a quick summary:
1. application developers should follow the official [go OpenTelemetry](https://pkg.go.dev/go.opentelemetry.io/otel) guidelines when instrumenting their applications.
2. node operators who want to configure OpenTelemetry exporters should set the `OTEL_EXPERIMENTAL_CONFIG_FILE` environment variable to the path of a yaml file which follows the OpenTelemetry declarative configuration format specified here: https://pkg.go.dev/go.opentelemetry.io/contrib/otelconf. As long as the `telemetry` package has been imported somwhere (it should already be imported if you are using the SDK), OpenTelemetry will be initialized automatically based on the configuration file.

NOTE: the go implementation of [otelconf](https://pkg.go.dev/go.opentelemetry.io/contrib/otelconf) is still under development and we will update our usage of it as it matures.