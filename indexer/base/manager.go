package indexerbase

type Manager struct {
	logger              Logger
	decoderResolver     DecoderResolver
	decoders            map[string]KVDecoder
	needLogicalDecoding bool
	listener            Listener
	listenerProcesses   []*listenerProcess
	done                chan struct{}
}

// ManagerOptions are the options for creating a new Manager.
type ManagerOptions struct {
	// DecoderResolver is the resolver for module decoders. It is required.
	DecoderResolver DecoderResolver

	// Listeners are the listeners that will be called when the manager receives events.
	Listeners []Listener

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

	done := opts.Done
	if done == nil {
		done = make(chan struct{})
	}

	mgr := &Manager{
		logger:              opts.Logger,
		decoderResolver:     opts.DecoderResolver,
		decoders:            map[string]KVDecoder{},
		needLogicalDecoding: false,
		listener:            Listener{},
		done:                done,
	}

	err := mgr.init(opts.Listeners)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

// Listener returns that listener that should be passed directly to the blockchain for managing
// all indexing.
func (p *Manager) init(listeners []Listener) error {
	// check each subscribed listener to see if we actually need to register the listener

	for _, listener := range listeners {
		if listener.StartBlock != nil {
			p.listener.StartBlock = p.startBlock
			break
		}
	}

	for _, listener := range listeners {
		if listener.OnBlockHeader != nil {
			p.listener.OnBlockHeader = p.onBlockHeader
			break
		}
	}

	for _, listener := range listeners {
		if listener.OnTx != nil {
			p.listener.OnTx = p.onTx
			break
		}
	}

	for _, listener := range listeners {
		if listener.OnEvent != nil {
			p.listener.OnEvent = p.onEvent
			break
		}
	}

	for _, listener := range listeners {
		if listener.Commit != nil {
			p.listener.Commit = p.commit
			break
		}
	}

	for _, listener := range listeners {
		if listener.OnObjectUpdate != nil {
			p.needLogicalDecoding = true
			p.listener.OnKVPair = p.onKVPair
			break
		}
	}

	if p.listener.OnKVPair == nil {
		for _, listener := range listeners {
			if listener.OnKVPair != nil {
				p.listener.OnKVPair = p.onKVPair
				break
			}
		}
	}

	if p.needLogicalDecoding {
		err := p.decoderResolver.Iterate(func(moduleName string, module ModuleDecoder) error {
			p.decoders[moduleName] = module.KVDecoder
			// TODO
			return nil
		})
		if err != nil {
			return err
		}
	}

	// initialize go routines for each listener
	for _, listener := range listeners {
		proc := &listenerProcess{
			listener:       listener,
			packetChan:     make(chan packet),
			commitDoneChan: make(chan error),
		}
		p.listenerProcesses = append(p.listenerProcesses, proc)

		// TODO initialize
		// TODO initialize module schema

		go proc.run()
	}

	return nil
}

func (p *Manager) startBlock(height uint64) error {
	if p.logger != nil {
		p.logger.Debug("start block", "height", height)
	}

	for _, proc := range p.listenerProcesses {
		if proc.listener.StartBlock != nil {
			continue
		}
		proc.packetChan <- packet{
			packetType: packetTypeStartBlock,
			data:       height,
		}
	}
	return nil
}

func (p *Manager) onBlockHeader(data BlockHeaderData) error {
	if p.logger != nil {
		p.logger.Debug("block header", "height", data.Height)
	}

	for _, proc := range p.listenerProcesses {
		if proc.listener.OnBlockHeader == nil {
			continue
		}
		proc.packetChan <- packet{
			packetType: packetTypeOnBlockHeader,
			data:       data,
		}
	}
	return nil
}

func (p *Manager) onTx(data TxData) error {
	if p.logger != nil {
		p.logger.Debug("tx", "txIndex", data.TxIndex)
	}

	for _, proc := range p.listenerProcesses {
		if proc.listener.OnTx == nil {
			continue
		}
		proc.packetChan <- packet{
			packetType: packetTypeOnTx,
			data:       data,
		}
	}
	return nil
}

func (p *Manager) onEvent(data EventData) error {
	if p.logger != nil {
		p.logger.Debug("event", "txIndex", data.TxIndex, "msgIndex", data.MsgIndex, "eventIndex", data.EventIndex)
	}

	for _, proc := range p.listenerProcesses {
		if proc.listener.OnEvent == nil {
			continue
		}
		proc.packetChan <- packet{
			packetType: packetTypeOnEvent,
			data:       data,
		}
	}

	return nil
}

func (p *Manager) commit() error {
	if p.logger != nil {
		p.logger.Debug("commit")
	}

	for _, proc := range p.listenerProcesses {
		if proc.listener.Commit == nil {
			continue
		}
		proc.packetChan <- packet{
			packetType: packetTypeCommit,
		}
	}

	// wait for all listeners to finish committing
	for _, proc := range p.listenerProcesses {
		err := <-proc.commitDoneChan
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Manager) onKVPair(data KVPairData) error {
	moduleName := data.ModuleName
	if p.logger != nil {
		p.logger.Debug("kv pair received", "moduleName", moduleName)
	}

	for _, proc := range p.listenerProcesses {
		if proc.listener.OnKVPair == nil {
			continue
		}
		proc.packetChan <- packet{
			packetType: packetTypeOnKVPair,
			data:       data,
		}
	}

	if !p.needLogicalDecoding {
		return nil
	}

	decoder, ok := p.decoders[moduleName]
	if !ok {
		// check for decoder when first seeing a module
		md, found, err := p.decoderResolver.LookupDecoder(moduleName)
		if err != nil {
			return err
		}
		if found {
			p.decoders[moduleName] = md.KVDecoder
			decoder = md.KVDecoder
		} else {
			p.decoders[moduleName] = nil
		}
	}

	if decoder == nil {
		return nil
	}

	update, handled, err := decoder(data.Key, data.Value, data.Delete)
	if err != nil {
		return err
	}
	if !handled {
		p.logger.Debug("not decoded", "moduleName", moduleName, "objectType", update.TypeName)
		return nil
	}

	p.logger.Debug("decoded",
		"moduleName", moduleName,
		"objectType", update.TypeName,
		"key", update.Key,
		"values", update.Value,
		"delete", update.Delete,
	)

	for _, proc := range p.listenerProcesses {
		if proc.listener.OnObjectUpdate == nil {
			continue
		}
		proc.packetChan <- packet{
			packetType: packetTypeOnObjectUpdate,
			data:       update,
		}
	}

	return nil
}
