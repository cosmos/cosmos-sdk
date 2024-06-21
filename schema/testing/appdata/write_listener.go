package appdatatest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"cosmossdk.io/schema/appdata"
)

func WriterListener(w io.Writer) appdata.Listener {
	return appdata.Listener{
		Initialize: func(ctx context.Context, data appdata.InitializationData) (lastBlockPersisted int64, err error) {

			_, err = fmt.Fprintf(w, "Initialize: %v\n", data)
			return 0, err
		},
		StartBlock: func(data appdata.StartBlockData) error {
			_, err := fmt.Fprintf(w, "StartBlock: %v\n", data)
			return err
		},
		OnTx:     nil,
		OnEvent:  nil,
		OnKVPair: nil,
		Commit: func() error {
			_, err := fmt.Fprintf(w, "Commit\n")
			return err
		},
		InitializeModuleData: func(data appdata.ModuleInitializationData) error {
			bz, err := json.Marshal(data)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "InitializeModuleData: %s\n", bz)
			return err
		},
		OnObjectUpdate: func(data appdata.ObjectUpdateData) error {
			bz, err := json.Marshal(data)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "OnObjectUpdate: %s\n", bz)
			return err
		},
	}
}
