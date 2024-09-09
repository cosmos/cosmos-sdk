package indexer

import (
	"fmt"

	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/decoding"
)

func addSyncAndSanityCheck(lastBlockPersisted uint64, listener appdata.Listener, mgrOpts IndexingOptions, moduleFilter ModuleFilterConfig) appdata.Listener {
	startBlock := listener.StartBlock
	initialized := false
	listener.StartBlock = func(data appdata.StartBlockData) error {
		if !initialized {
			if err := doSyncAndSanityCheck(lastBlockPersisted, data, listener, mgrOpts, moduleFilter); err != nil {
				return err
			}
			initialized = true
		}
		if startBlock != nil {
			return startBlock(data)
		}
		return nil
	}
	return listener
}

func doSyncAndSanityCheck(lastBlockPersisted uint64, data appdata.StartBlockData, listener appdata.Listener, mgrOpts IndexingOptions, moduleFilter ModuleFilterConfig) error {
	if lastBlockPersisted == 0 {
		if data.Height == 1 {
			// this is the first block anyway so nothing to sync
			return nil
		}
		// need to do a catch-up sync
		return decoding.Sync(listener, mgrOpts.SyncSource, mgrOpts.Resolver, decoding.SyncOptions{
			ModuleFilter: moduleFilter.ToFunction(),
		})
	} else if lastBlockPersisted+1 != data.Height {
		// we are out of sync so through an error
		return fmt.Errorf("fatal error: indexer is out of sync, last block persisted: %d, current block height: %d", lastBlockPersisted, data.Height)
	}
	// all good
	return nil
}
