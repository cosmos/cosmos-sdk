// Package telemetry provides instrumentation tools for other modules and packages.
package telemetry

import "time"

const (
	//Default Values of configuration

	//defaultEnabled is the default value for the enable config.
	defaultEnabled = false
	//defaultEnableHostname  is the default value for the enable hostname  config.
	defaultEnableHostname = false

	//defaultEnableHostnameLabel is the default value for the enable hostname label  config.
	defaultEnableHostnameLabel = false

	//defaultEnableServiceLabel is the default value for the enable hostname label config.
	defaultEnableServiceLabel = false

	//defaultUseGlobalMetricRegistration is the default value for the global registration for go-metrics.
	defaultUseGlobalMetricRegistration = true
)

var (
	//defaultPrometheusRetentionTime is the default value for the enable prometheus retention time  config.
	defaultPrometheusRetentionTime = 60 * time.Second
)
