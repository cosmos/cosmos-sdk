package core

// RuntimeCompatibilityVersion indicates what semantic version of the runtime
// module is required to support the full API exposed in this version of core.
// The runtime module that is loaded can emit an error or warning if the version
// specified here is greater than what that runtime can support and some
// features may not work as expected.
const RuntimeCompatibilityVersion = 1
