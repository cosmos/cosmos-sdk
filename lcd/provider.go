package lcd

// Provider is used to get more validators by other means.
//
// Examples: MemProvider, files.Provider, client.Provider, CacheProvider....
type Provider interface {
	// StoreCommit saves a FullCommit after we have verified it,
	// so we can query for it later. Important for updating our
	// store of trusted commits.
	StoreCommit(fc FullCommit) error
	// GetByHeight returns the closest commit with height <= h.
	GetByHeight(h int64) (FullCommit, error)
	// GetByHash returns a commit exactly matching this validator hash.
	GetByHash(hash []byte) (FullCommit, error)
	// LatestCommit returns the newest commit stored.
	LatestCommit() (FullCommit, error)
}

// cacheProvider allows you to place one or more caches in front of a source
// Provider.  It runs through them in order until a match is found.
// So you can keep a local cache, and check with the network if
// no data is there.
type cacheProvider struct {
	Providers []Provider
}

// NewCacheProvider returns a new provider which wraps multiple other providers.
func NewCacheProvider(providers ...Provider) Provider {
	return cacheProvider{
		Providers: providers,
	}
}

// StoreCommit tries to add the seed to all providers.
//
// Aborts on first error it encounters (closest provider)
func (c cacheProvider) StoreCommit(fc FullCommit) (err error) {
	for _, p := range c.Providers {
		err = p.StoreCommit(fc)
		if err != nil {
			break
		}
	}
	return err
}

// GetByHeight should return the closest possible match from all providers.
//
// The Cache is usually organized in order from cheapest call (memory)
// to most expensive calls (disk/network). However, since GetByHeight returns
// a FullCommit at h' <= h, if the memory has a seed at h-10, but the network would
// give us the exact match, a naive "stop at first non-error" would hide
// the actual desired results.
//
// Thus, we query each provider in order until we find an exact match
// or we finished querying them all.  If at least one returned a non-error,
// then this returns the best match (minimum h-h').
func (c cacheProvider) GetByHeight(h int64) (fc FullCommit, err error) {
	for _, p := range c.Providers {
		var tfc FullCommit
		tfc, err = p.GetByHeight(h)
		if err == nil {
			if tfc.Height() > fc.Height() {
				fc = tfc
			}
			if tfc.Height() == h {
				break
			}
		}
	}
	// even if the last one had an error, if any was a match, this is good
	if fc.Height() > 0 {
		err = nil
	}
	return fc, err
}

// GetByHash returns the FullCommit for the hash or an error if the commit is not found.
func (c cacheProvider) GetByHash(hash []byte) (fc FullCommit, err error) {
	for _, p := range c.Providers {
		fc, err = p.GetByHash(hash)
		if err == nil {
			break
		}
	}
	return fc, err
}

// LatestCommit returns the latest FullCommit or an error if no commit exists.
func (c cacheProvider) LatestCommit() (fc FullCommit, err error) {
	for _, p := range c.Providers {
		var tfc FullCommit
		tfc, err = p.LatestCommit()
		if err == nil && tfc.Height() > fc.Height() {
			fc = tfc
		}
	}
	// even if the last one had an error, if any was a match, this is good
	if fc.Height() > 0 {
		err = nil
	}
	return fc, err
}
