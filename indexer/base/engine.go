package indexerbase

import "fmt"

type Engine struct {
	logger              Logger
	decoderResolver     DecoderResolver
	tables              map[string]Table
	decoders            map[string]KVDecoder
	logicalListeners    []LogicalListener
	physicalListeners   []PhysicalListener
	needLogicalDecoding bool
}

type EngineOptions struct {
	ModuleDecoders    DecoderResolver
	LogicalListeners  []LogicalListener
	PhysicalListeners []PhysicalListener
	Logger            Logger
}

func NewEngine(opts EngineOptions) (*Engine, error) {
	if opts.Logger != nil {
		opts.Logger.Info("Initializing indexer engine")
	}

	schema := Schema{}
	tables := map[string]Table{}
	decoders := map[string]KVDecoder{}
	for moduleName, module := range opts.ModuleDecoders {
		modSchema := module.Schema
		for _, table := range modSchema.Tables {
			table.Name = moduleName + "_" + table.Name
			if _, ok := tables[table.Name]; ok {
				return nil, fmt.Errorf("duplicate table name: %s", table.Name)
			}
			tables[table.Name] = table
			schema.Tables = append(schema.Tables, table)
		}
		decoders[moduleName] = module.KVDecoder
	}

	var physicalListeners []PhysicalListener
	physicalListeners = append(physicalListeners, opts.PhysicalListeners...)

	logicalSetupData := LogicalSetupData{Schema: schema}
	for _, indexer := range opts.LogicalListeners {
		if indexer.EnsureSetup != nil {
			if err := indexer.EnsureSetup(logicalSetupData); err != nil {
				return nil, err
			}
		}

		physicalListeners = append(physicalListeners, indexer.PhysicalListener)
	}

	return &Engine{
		logger:            opts.Logger,
		tables:            tables,
		decoders:          decoders,
		physicalListeners: physicalListeners,
		logicalListeners:  opts.LogicalListeners,
	}, nil
}

func (p *Engine) PhysicalListener() PhysicalListener {
	l := PhysicalListener{}

	// check each subscribed listener to see if we actually need to register the listener

	for _, listener := range p.physicalListeners {
		if listener.StartBlock != nil {
			l.StartBlock = p.startBlock
			break
		}
	}

	for _, listener := range p.physicalListeners {
		if listener.OnBlockHeader != nil {
			l.OnBlockHeader = p.onBlockHeader
			break
		}
	}

	for _, listener := range p.physicalListeners {
		if listener.OnTx != nil {
			l.OnTx = p.onTx
			break
		}
	}

	for _, listener := range p.physicalListeners {
		if listener.OnEvent != nil {
			l.OnEvent = p.onEvent
			break
		}
	}

	for _, listener := range p.physicalListeners {
		if listener.Commit != nil {
			l.Commit = p.commit
			break
		}
	}

	for _, listener := range p.logicalListeners {
		if listener.OnEntityUpdate != nil {
			p.needLogicalDecoding = true
			l.OnKVPair = p.onKVPair
			break
		}
	}

	if l.OnKVPair == nil {
		for _, listener := range p.physicalListeners {
			if listener.OnKVPair != nil {
				l.OnKVPair = p.onKVPair
				break
			}
		}
	}

	return l
}

func (p *Engine) startBlock(height uint64) error {
	if p.logger != nil {
		p.logger.Debug("start block", "height", height)
	}

	for _, listener := range p.physicalListeners {
		if listener.StartBlock == nil {
			continue
		}
		if err := listener.StartBlock(height); err != nil {
			return err
		}
	}
	return nil
}

func (p *Engine) onBlockHeader(data BlockHeaderData) error {
	if p.logger != nil {
		p.logger.Debug("block header", "height", data.Height)
	}

	for _, listener := range p.physicalListeners {
		if listener.OnBlockHeader == nil {
			continue
		}
		if err := listener.OnBlockHeader(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *Engine) onTx(data TxData) error {
	for _, listener := range p.physicalListeners {
		if listener.OnTx == nil {
			continue
		}
		if err := listener.OnTx(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *Engine) onEvent(data EventData) error {
	for _, listener := range p.physicalListeners {
		if err := listener.OnEvent(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *Engine) commit() error {
	if p.logger != nil {
		p.logger.Debug("commit")
	}

	for _, listener := range p.physicalListeners {
		if err := listener.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Engine) onKVPair(storeKey string, key, value []byte, delete bool) error {
	if p.logger != nil {
		p.logger.Debug("kv pair", "storeKey", storeKey, "delete", delete)
	}

	for _, listener := range p.physicalListeners {
		if listener.OnKVPair == nil {
			continue
		}
		if err := listener.OnKVPair(storeKey, key, value, delete); err != nil {
			return err
		}
	}

	if !p.needLogicalDecoding {
		return nil
	}

	decoder, ok := p.decoders[storeKey]
	if !ok {
		return nil
	}

	update, handled, err := decoder(key, value)
	if err != nil {
		return err
	}
	if !handled {
		p.logger.Info("not decoded", "storeKey", storeKey, "tableName", update.TableName)
		return nil
	}

	p.logger.Info("decoded",
		"storeKey", storeKey,
		"tableName", update.TableName,
		"key", update.Key,
		"values", update.Value,
		"delete", update.Delete,
	)

	// prepend module name to table name
	update.TableName = fmt.Sprintf("%s_%s", storeKey, update.TableName)

	for _, indexer := range p.logicalListeners {
		if indexer.OnEntityUpdate == nil {
			continue
		}
		if err := indexer.OnEntityUpdate(update); err != nil {
			return err
		}
	}

	return nil
}
