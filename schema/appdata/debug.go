package appdata

import (
	"fmt"
	"io"
)

func DebugListener(out io.Writer) Listener {
	res := packetForwarder(func(p Packet) error {
		_, err := fmt.Fprintln(out, p)
		return err
	})
	//res.Initialize = func(ctx context.Context, data InitializationData) (lastBlockPersisted int64, err error) {
	//	_, err = fmt.Fprintf(out, "Initialize: %v\n", data)
	//	return 0, err
	//}
	return res
}

func packetForwarder(f func(Packet) error) Listener {
	return Listener{
		//Initialize:           nil, // can't be forwarded
		InitializeModuleData: func(data ModuleInitializationData) error { return f(data) },
		OnTx:                 func(data TxData) error { return f(data) },
		OnEvent:              func(data EventData) error { return f(data) },
		OnKVPair:             func(data KVPairData) error { return f(data) },
		OnObjectUpdate:       func(data ObjectUpdateData) error { return f(data) },
		StartBlock:           func(data StartBlockData) error { return f(data) },
		Commit:               func(data CommitData) error { return f(data) },
	}
}
