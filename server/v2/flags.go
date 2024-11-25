// Package serverv2 defines constants for server configuration flags and output formats.
package serverv2

import "fmt"

// start flags are prefixed with the server name
// this allows viper to properly bind the flags
func prefix(f string) string {
	return fmt.Sprintf("%s.%s", serverName, f)
}

var (
	FlagMinGasPrices = prefix("minimum-gas-prices")
	FlagCPUProfiling = prefix("cpu-profile")
)

const (
	// FlagHome specifies the home directory flag.
	FlagHome = "home"

	FlagLogLevel   = "log_level"    // Sets the logging level
	FlagLogFormat  = "log_format"   // Specifies the log output format
	FlagLogNoColor = "log_no_color" // Disables colored log output
	FlagTrace      = "trace"        // Enables trace-level logging

	// OutputFormatJSON defines the JSON output format option.
	OutputFormatJSON = "json"
)
