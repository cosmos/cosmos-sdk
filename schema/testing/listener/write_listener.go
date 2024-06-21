package listenertest

import (
	"encoding/json"
	"fmt"
	"io"

	"cosmossdk.io/schema"
)

func WriterListener(w io.Writer) blockdata.Listener {
	return blockdata.Listener{
		Initialize: func(data blockdata.InitializationData) (lastBlockPersisted int64, err error) {

			_, err = fmt.Fprintf(w, "Initialize: %v\n", data)
			return 0, err
		},
		StartBlock: func(u uint64) error {
			_, err := fmt.Fprintf(w, "StartBlock: %d\n", u)
			return err
		},
		OnBlockHeader: func(data blockdata.BlockHeaderData) error {
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
		InitializeModuleData: func(moduleName string, schema schema.ModuleSchema) error {
			bz, err := json.Marshal(schema)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "InitializeModuleData: %s %s\n", moduleName, bz)
			return err
		},
		OnObjectUpdate: func(data blockdata.ObjectUpdateData) error {
			bz, err := json.Marshal(data)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "OnObjectUpdate: %s\n", bz)
			return err
		},
	}
}
