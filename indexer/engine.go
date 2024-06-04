package indexer

type Engine struct {
	moduleDecoders map[string][]ModuleStateDecoder
	indexers       []Indexer
}

type EngineOptions[ModuleT any] struct {
	ModuleSet map[string]ModuleT
	Decoders  []Decoder
	Indexers  []Indexer
}

func NewEngine[T any](opts EngineOptions[T]) *Engine {
	moduleDecoders := make(map[string][]ModuleStateDecoder)
	for moduleName, module := range opts.ModuleSet {
		for _, decoder := range opts.Decoders {
			modDecoder := decoder.ExtractModuleDecoder(moduleName, module)
			if modDecoder != nil {
				existing, ok := moduleDecoders[moduleName]
				if !ok {
					moduleDecoders[moduleName] = []ModuleStateDecoder{modDecoder}
				} else {
					moduleDecoders[moduleName] = append(existing, modDecoder)
				}
			}
		}

	}
	return &Engine{
		moduleDecoders: moduleDecoders,
		indexers:       opts.Indexers,
	}
}

func (p *Engine) StartBlock(data *BlockHeaderData) error {
	for _, indexer := range p.indexers {
		if err := indexer.StartBlock(data.Height); err != nil {
			return err
		}
		if err := indexer.IndexBlockHeader(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *Engine) ReceiveTx(data *TxData) error {
	for _, indexer := range p.indexers {
		if err := indexer.IndexTx(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *Engine) ReceiveEvent(data *EventData) error {
	for _, indexer := range p.indexers {
		if err := indexer.IndexEvent(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *Engine) ReceiveStateSet(storeKey string, key, value []byte) error {
	decoders := p.moduleDecoders[storeKey]
	if decoders == nil {
		return nil
	}
	for _, decoder := range decoders {
		update, err := decoder.DecodeSet(key, value)
		if err != nil {
			return err
		}
		if update == nil {
			continue
		}
		for _, indexer := range p.indexers {
			if err := indexer.IndexEntityUpdate(update); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Engine) ReceiveStateDelete(storeKey string, key []byte, prune bool) error {
	decoders := p.moduleDecoders[storeKey]
	if decoders == nil {
		return nil
	}
	for _, decoder := range decoders {
		del, err := decoder.DecodeDelete(key)
		if err != nil {
			return err
		}
		if del == nil {
			continue
		}
		for _, indexer := range p.indexers {
			if err := indexer.IndexEntityDelete(del); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Engine) Commit() error {
	for _, indexer := range p.indexers {
		if err := indexer.Commit(); err != nil {
			return err
		}
	}
	return nil
}
