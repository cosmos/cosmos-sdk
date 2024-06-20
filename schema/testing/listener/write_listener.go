package listenertest

import (
	"encoding/json"
	"fmt"
	"io"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/listener"
)

func WriterListener(w io.Writer) listener.Listener {
	return listener.Listener{
		Initialize: func(data listener.InitializationData) (lastBlockPersisted int64, err error) {

			_, err = fmt.Fprintf(w, "Initialize: %v\n", data)
			return 0, err
		},
		StartBlock: func(u uint64) error {
			_, err := fmt.Fprintf(w, "StartBlock: %d\n", u)
			return err
		},
		OnBlockHeader: func(data listener.BlockHeaderData) error {
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
		InitializeModuleSchema: func(moduleName string, schema schema.ModuleSchema) error {
			bz, err := json.Marshal(schema)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "InitializeModuleSchema: %s %s\n", moduleName, bz)
			return err
		},
		OnObjectUpdate: func(data listener.ObjectUpdateData) error {
			bz, err := json.Marshal(data)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "OnObjectUpdate: %s\n", bz)
			return err
		},
	}
}
