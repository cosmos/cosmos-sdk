package blockdata

type Packet interface {
	apply(*Listener) error
}

func (l Listener) SendPacket(p Packet) error {
	return p.apply(&l)
}

func (m ModuleInitializationData) apply(l *Listener) error {
	if l.InitializeModuleData == nil {
		return nil
	}
	return l.InitializeModuleData(m)
}

func (b StartBlockData) apply(l *Listener) error {
	if l.StartBlock == nil {
		return nil
	}
	return l.StartBlock(b)
}

func (t TxData) apply(l *Listener) error {
	if l.OnTx == nil {
		return nil
	}
	return l.OnTx(t)
}

func (e EventData) apply(l *Listener) error {
	if l.OnEvent == nil {
		return nil
	}
	return l.OnEvent(e)
}

func (k KVPairData) apply(l *Listener) error {
	if l.OnKVPair == nil {
		return nil
	}
	return l.OnKVPair(k)
}

func (o ObjectUpdateData) apply(l *Listener) error {
	if l.OnObjectUpdate == nil {
		return nil
	}
	return l.OnObjectUpdate(o)
}

func (c Commit) apply(l *Listener) error {
	if l.Commit == nil {
		return nil
	}
	return l.Commit()
}
