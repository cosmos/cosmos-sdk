package indexer

import (
	"cosmossdk.io/schema"
	"cosmossdk.io/schema/listener"
)

type Manager struct {
	logger              Logger
	decoderResolver     DecoderResolver
	decoders            map[string]schema.KVDecoder
	needLogicalDecoding bool
	listener            listener.Listener
	initialized         bool
	commitChan          chan error
}

// ManagerOptions are the options for creating a new Manager.
type ManagerOptions struct {
	// DecoderResolver is the resolver for module decoders. It is required.
	DecoderResolver DecoderResolver

	// Listeners are the listeners that will be called when the manager receives events.
	Listeners []listener.Listener

	// SyncSource is the source that will be used do initial indexing of modules with pre-existing
	// state. It is optional, but if it is not provided, indexing can only be starting when a node
	// is synced from genesis.
	SyncSource SyncSource

	// Logger is the logger that will be used by the manager. It is optional.
	Logger Logger

	BufferSize int

	Done chan struct{}
}

// NewManager creates a new Manager with the provided options.
func NewManager(opts ManagerOptions) (*Manager, error) {
	if opts.Logger != nil {
		opts.Logger.Info("Initializing indexer manager")
	}

	mgr := &Manager{
		logger:              opts.Logger,
		decoderResolver:     opts.DecoderResolver,
		decoders:            map[string]schema.KVDecoder{},
		needLogicalDecoding: false,
		listener:            listener.AsyncMultiplex(opts.Listeners, opts.BufferSize),
		commitChan:          make(chan error),
	}

	return mgr, nil
}

//// Listener returns that listener that should be passed directly to the blockchain for managing
//// all indexing.
//func (p *Manager) setup(listeners []listener.Listener) error {
//	// check each subscribed listener to see if we actually need to register the listener
//
//	p.listener.Initialize = p.initialize
//
//	for _, l := range listeners {
//		if l.StartBlock != nil {
//			p.listener.StartBlock = p.startBlock
//			break
//		}
//	}
//
//	for _, l := range listeners {
//		if l.OnBlockHeader != nil {
//			p.listener.OnBlockHeader = p.onBlockHeader
//			break
//		}
//	}
//
//	for _, l := range listeners {
//		if l.OnTx != nil {
//			p.listener.OnTx = p.onTx
//			break
//		}
//	}
//
//	for _, listener := range listeners {
//		if listener.OnEvent != nil {
//			p.listener.OnEvent = p.onEvent
//			break
//		}
//	}
//
//	for _, listener := range listeners {
//		if listener.Commit != nil {
//			p.listener.Commit = p.commit
//			break
//		}
//	}
//
//	for _, listener := range listeners {
//		if listener.OnObjectUpdate != nil {
//			p.needLogicalDecoding = true
//			p.listener.OnKVPair = p.onKVPair
//			break
//		}
//	}
//
//	if p.listener.OnKVPair == nil {
//		for _, listener := range listeners {
//			if listener.OnKVPair != nil {
//				p.listener.OnKVPair = p.onKVPair
//				break
//			}
//		}
//	}
//
//	if p.needLogicalDecoding {
//		p.listener.InitializeModuleSchema = p.initializeModuleSchema
//		p.listener.OnObjectUpdate = p.onObjectUpdate
//	}
//
//	// initialize go routines for each listener
//	for _, l := range listeners {
//		proc := &listenerProcess{
//			listener:       l,
//			packetChan:     make(chan listener.Packet),
//			commitDoneChan: make(chan error),
//		}
//		p.listenerProcesses = append(p.listenerProcesses, proc)
//	}
//
//	return nil
//}
//
//func (p *Manager) initialize(data listener.InitializationData) (lastBlockPersisted int64, err error) {
//	if p.logger != nil {
//		p.logger.Debug("initialize")
//	}
//
//	// setup logical decoding
//	if p.needLogicalDecoding {
//		err = p.decoderResolver.Iterate(func(moduleName string, module schema.ModuleCodec) error {
//			// if the schema was already initialized by the data source by InitializeModuleSchema,
//			// then this is an error
//			if _, ok := p.decoders[moduleName]; ok {
//				return fmt.Errorf("module schema for %s already initialized", moduleName)
//			}
//
//			p.decoders[moduleName] = module.KVDecoder
//
//			for _, proc := range p.listenerProcesses {
//				if proc.listener.InitializeModuleSchema == nil {
//					continue
//				}
//				err := proc.listener.InitializeModuleSchema(moduleName, module.Schema)
//				if err != nil {
//					return err
//				}
//			}
//
//			return nil
//		})
//		if err != nil {
//			return
//		}
//	}
//
//	// call initialize
//	for _, proc := range p.listenerProcesses {
//		if proc.listener.Initialize == nil {
//			continue
//		}
//		lastBlockPersisted, err = proc.listener.Initialize(data)
//		if err != nil {
//			return
//		}
//	}
//
//	// start go routines
//	for _, proc := range p.listenerProcesses {
//		go proc.run()
//	}
//
//	p.initialized = true
//
//	return
//}
//
//func (p *Manager) startBlock(height uint64) error {
//	if p.logger != nil {
//		p.logger.Debug("start block", "height", height)
//	}
//
//	for _, proc := range p.listenerProcesses {
//		if proc.listener.StartBlock != nil {
//			continue
//		}
//		proc.packetChan <- listener.StartBlock(height)
//	}
//	return nil
//}
//
//func (p *Manager) onBlockHeader(data listener.BlockHeaderData) error {
//	if p.logger != nil {
//		p.logger.Debug("block header", "height", data.Height)
//	}
//
//	for _, proc := range p.listenerProcesses {
//		if proc.listener.OnBlockHeader == nil {
//			continue
//		}
//		proc.packetChan <- data
//	}
//	return nil
//}
//
//func (p *Manager) onTx(data listener.TxData) error {
//	if p.logger != nil {
//		p.logger.Debug("tx", "txIndex", data.TxIndex)
//	}
//
//	for _, proc := range p.listenerProcesses {
//		if proc.listener.OnTx == nil {
//			continue
//		}
//		proc.packetChan <- data
//	}
//	return nil
//}
//
//func (p *Manager) onEvent(data listener.EventData) error {
//	if p.logger != nil {
//		p.logger.Debug("event", "txIndex", data.TxIndex, "msgIndex", data.MsgIndex, "eventIndex", data.EventIndex)
//	}
//
//	for _, proc := range p.listenerProcesses {
//		if proc.listener.OnEvent == nil {
//			continue
//		}
//		proc.packetChan <- data
//	}
//
//	return nil
//}
//
//func (p *Manager) commit() error {
//	if p.logger != nil {
//		p.logger.Debug("commit")
//	}
//
//	for _, proc := range p.listenerProcesses {
//		if proc.listener.Commit == nil {
//			continue
//		}
//		proc.packetChan <- listener.Commit{}
//	}
//
//	// wait for all listeners to finish committing
//	for _, proc := range p.listenerProcesses {
//		err := <-proc.commitDoneChan
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (p *Manager) onKVPair(data listener.KVPairData) error {
//	moduleName := data.ModuleName
//	if p.logger != nil {
//		p.logger.Debug("kv pair received", "moduleName", moduleName)
//	}
//
//	for _, proc := range p.listenerProcesses {
//		if proc.listener.OnKVPair == nil {
//			continue
//		}
//		proc.packetChan <- data
//	}
//
//	if !p.needLogicalDecoding {
//		return nil
//	}
//
//	decoder, ok := p.decoders[moduleName]
//	if !ok {
//		// check for decoder when first seeing a module
//		md, found, err := p.decoderResolver.LookupDecoder(moduleName)
//		if err != nil {
//			return err
//		}
//		if found {
//			p.decoders[moduleName] = md.KVDecoder
//			decoder = md.KVDecoder
//		} else {
//			p.decoders[moduleName] = nil
//		}
//	}
//
//	if decoder == nil {
//		return nil
//	}
//
//	update, handled, err := decoder(data.Key, data.Value, data.Delete)
//	if err != nil {
//		return err
//	}
//	if !handled {
//		p.logger.Debug("not decoded", "moduleName", moduleName, "objectType", update.TypeName)
//		return nil
//	}
//
//	p.logger.Debug("decoded",
//		"moduleName", moduleName,
//		"objectType", update.TypeName,
//		"key", update.Key,
//		"values", update.Value,
//		"delete", update.Delete,
//	)
//
//	for _, proc := range p.listenerProcesses {
//		if proc.listener.OnObjectUpdate == nil {
//			continue
//		}
//		proc.packetChan <- data
//	}
//
//	return nil
//}
//
//func (p *Manager) initializeModuleSchema(module string, schema schema.ModuleSchema) error {
//	if p.initialized {
//		return fmt.Errorf("cannot initialize module schema after initialization")
//	}
//
//	for _, proc := range p.listenerProcesses {
//		// set the decoder for the module to so that we know that it is already initialized,
//		// but that also we are not handling decoding - in this case the data source
//		// should be doing the decoding and passing it to the manager in OnObjectUpdate
//		p.decoders[module] = nil
//
//		if proc.listener.InitializeModuleSchema == nil {
//			continue
//		}
//		err := proc.listener.InitializeModuleSchema(module, schema)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (p *Manager) onObjectUpdate(data listener.ObjectUpdateData) error {
//	for _, proc := range p.listenerProcesses {
//		if proc.listener.OnObjectUpdate == nil {
//			continue
//		}
//		proc.packetChan <- data
//	}
//	return nil
//}
