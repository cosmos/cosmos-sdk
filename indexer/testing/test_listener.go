package indexertesting

import (
	"fmt"
	"io"
	"os"

	indexerbase "cosmossdk.io/indexer/base"
)

func StdoutListener() indexerbase.Listener {
	return WriterListener(os.Stdout)
}

func WriterListener(w io.Writer) indexerbase.Listener {
	return indexerbase.Listener{
		Initialize: func(data indexerbase.InitializationData) (lastBlockPersisted int64, err error) {
			_, err = fmt.Fprintf(w, "Initialize: %v\n", data)
			return 0, err
		},
		StartBlock: func(u uint64) error {
			_, err := fmt.Fprintf(w, "StartBlock: %d\n", u)
			return err
		},
		OnBlockHeader: func(data indexerbase.BlockHeaderData) error {
			_, err := fmt.Fprintf(w, "OnBlockHeader: %v\n", data)
			return err
		},
		OnTx:     nil,
		OnEvent:  nil,
		OnKVPair: nil,
		Commit: func() error {
			_, err := fmt.Fprintf(w, "Commit\n")
			return err
		},
		InitializeModuleSchema: func(moduleName string, schema indexerbase.ModuleSchema) error {
			_, err := fmt.Fprintf(w, "InitializeModuleSchema: %s %v\n", moduleName, schema)
			return err
		},
		OnObjectUpdate: func(moduleName string, data indexerbase.ObjectUpdate) error {
			//_, err := fmt.Fprintf(w, "OnObjectUpdate: %s: %v\n", moduleName, data)
			//return err
			return nil
		},
		CommitCatchupSync: nil,
	}
}
