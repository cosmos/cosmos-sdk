//go:build consensus_break_test

package baseapp

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// injectConsensusBreak is a test helper function to inject non-determinism
// into the app hash, causing a consensus failure at configurable intervals.
// This function is only included in builds with the 'consensus_break_test' tag.
// Environment variables: ENABLE_CONSENSUS_BREAK=true, CONSENSUS_BREAK_INTERVAL=N (default: 10)
func (app *BaseApp) injectConsensusBreak(appHash []byte, height int64, path string) []byte {
	if os.Getenv("ENABLE_CONSENSUS_BREAK") != "true" {
		return appHash
	}

	interval := int64(10)
	if envInterval := os.Getenv("CONSENSUS_BREAK_INTERVAL"); envInterval != "" {
		if parsed, err := strconv.ParseInt(envInterval, 10, 64); err == nil && parsed > 0 {
			interval = parsed
		}
	}

	if height > 0 && height%interval == 0 {
		timeNano := time.Now().UnixNano()
		timeBytes := []byte(fmt.Sprintf("%d", timeNano))
		modifiedHash := append(appHash, timeBytes...)

		app.logger.Error(
			"consensus break injected for testing",
			"path", path,
			"height", height,
			"interval", interval,
			"original_hash", fmt.Sprintf("%X", appHash),
			"modified_hash", fmt.Sprintf("%X", modifiedHash),
			"time_injected", timeNano,
		)

		return modifiedHash
	}

	return appHash
}
