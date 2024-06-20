package listener

type Packet interface {
	apply(*Listener) error
}

func (l Listener) ApplyPacket(p Packet) error {
	return p.apply(&l)
}

func PacketCollector(onPacket func(Packet) error) Listener {
	return Listener{
		StartBlock:     func(u uint64) error { return onPacket(StartBlock(u)) },
		OnBlockHeader:  func(data BlockHeaderData) error { return onPacket(data) },
		OnTx:           func(data TxData) error { return onPacket(data) },
		OnEvent:        func(data EventData) error { return onPacket(data) },
		OnKVPair:       func(data KVPairData) error { return onPacket(data) },
		Commit:         func() error { return onPacket(Commit{}) },
		OnObjectUpdate: func(data ObjectUpdateData) error { return onPacket(data) },
	}
}
