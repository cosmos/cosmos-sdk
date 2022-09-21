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
)

var (
	//defaultPrometheusRetentionTime is the default value for the enable prometheus retention time  config.
	defaultPrometheusRetentionTime = 1 * time.Second
)
