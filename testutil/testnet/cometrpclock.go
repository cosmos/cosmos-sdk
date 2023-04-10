package testnet

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/cometbft/cometbft/node"
)

// CometBFT v0.37 uses a singleton to manage the RPC "environment".
// v0.38 will not have that restriction, which was removed in commit:
// https://github.com/cometbft/cometbft/commit/3324f49fb7e7b40189726746493e83b82a61b558
//
// We manage a corresponding global lock to ensure
// we don't attempt to use multiple active RPC servers in one process,
// which would result in unpredictable or incorrect behavior.
// Once the SDK adopts Comet v0.38+, we can remove this global lock mechanism.

// Our singleton complementing Comet's global RPC state.
var globalCometMu = new(cometRPCMutex)

type cometRPCMutex struct {
	mu sync.Mutex

	prevLockStack []byte
}

// CometRPCInUseError is returned on a failure to acquire
// the global comet RPC lock.
//
// This type will be removed once the Cosmos SDK adopts CometBFT v0.38 or newer.
type CometRPCInUseError struct {
	prevStack []byte
}

func (e CometRPCInUseError) Error() string {
	return fmt.Sprintf(`Failed to acquire global lock for Comet RPC servers.

If this in a test using t.Parallel(), remove the call to t.Parallel().

If this is in a test NOT using t.Parallel,
ensure that other callers call both Stop() and Wait() on the nodes.

If there are multiple comet instances in one test using RPC servers,
ensure that only one instance has the RPC listener enabled.

These restrictions will be loosened once cosmos-sdk adopts comet-bft v0.38 or newer.

Stack where lock was previously acquired:
%s
`, e.prevStack)
}

// Acquire attempts to acquire the underlying mutex.
// If it cannot be acquired on the first attempt,
// Acquire returns a [CometRPCInUseError] value.
func (m *cometRPCMutex) Acquire() error {
	if !m.mu.TryLock() {
		// If we can't acquire the lock,
		// there is another active comet node using RPC.
		//
		// This was initially going to be a panic,
		// but we can't easily write tests against that since
		// the panic occurs in a separate goroutine
		// when called through NewNetwork.
		//
		// Note, reading m.prevLockStack without holding m.mu
		// is technically a data race,
		// since it is possible that the previous caller was about to unlock.
		// Nonetheless, the programmer is responsible for avoiding that situation,
		// and a data race during a failure isn't particularly concerning.
		return CometRPCInUseError{prevStack: m.prevLockStack}
	}

	// Now we hold the lock, so first record the stack when the lock was taken.
	m.prevLockStack = debug.Stack()

	return nil
}

// Release unlocks m depending on n.
// If n is nil, m is unlocked immediately.
// If n is not nil, a new goroutine is created
// and n is released after the node has finished running.
func (m *cometRPCMutex) Release(n *node.Node) {
	if n == nil {
		m.prevLockStack = nil
		m.mu.Unlock()
		return
	}

	go m.releaseAfterWait(n)
}

func (m *cometRPCMutex) releaseAfterWait(n *node.Node) {
	n.Wait()
	m.prevLockStack = nil
	m.mu.Unlock()
}
