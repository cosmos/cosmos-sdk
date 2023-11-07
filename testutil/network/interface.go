package network

import "time"

// NetworkI is an interface for a network of validators.
// It is used to abstract over the different network types (in-process, docker, etc.).
// if used there is a requirement to expose query and tx client for the nodes
type NetworkI interface {
	// GetValidators returns the validators in the network
	GetValidators() []ValidatorI
	// WaitForHeight waits for the network to reach the given height
	WaitForNextBlock() error
	// WaitForHeight waits for the network to reach the given height
	WaitForHeight(height int64) (int64, error)
	// WaitForHeightWithTimeout waits for the network to reach the given height or times out
	WaitForHeightWithTimeout(int64, time.Duration) (int64, error)
	// RetryForHeight retries the given function until it returns no error or the given number of blocks has passed
	RetryForBlocks(retryFunc func() error, blocks int) error
	// LatestHeight returns the latest height of the network
	LatestHeight() (int64, error)

	Cleanup()
}
