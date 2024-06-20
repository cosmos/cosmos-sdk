package indexer

import (
	"fmt"

	"cosmossdk.io/schema/listener"
)

type packetType int

const (
	packetTypeStartBlock = iota
	packetTypeOnBlockHeader
	packetTypeOnTx
	packetTypeOnEvent
	packetTypeOnKVPair
	packetTypeOnObjectUpdate
	packetTypeCommit
)

type packet struct {
	packetType packetType
	data       interface{}
}

type listenerProcess struct {
	listener       listener.Listener
	packetChan     chan packet
	err            error
	commitDoneChan chan error
	cancel         chan struct{}
}

func (l *listenerProcess) run() {
	for {
		select {
		case packet := <-l.packetChan:
			if l.processPacket(packet) {
				return // stop processing packets
			}
		case <-l.cancel:
			return
		}
	}
}

func (l *listenerProcess) processPacket(p packet) bool {
	if l.err != nil {
		if p.packetType == packetTypeCommit {
			l.commitDoneChan <- l.err
			return true
		}
		return false
	}

	switch p.packetType {
	case packetTypeStartBlock:
		if l.listener.StartBlock != nil {
			l.err = l.listener.StartBlock(p.data.(uint64))
		}
	case packetTypeOnBlockHeader:
		if l.listener.OnBlockHeader != nil {
			l.err = l.listener.OnBlockHeader(p.data.(listener.BlockHeaderData))
		}
	case packetTypeOnTx:
		if l.listener.OnTx != nil {
			l.err = l.listener.OnTx(p.data.(listener.TxData))
		}
	case packetTypeOnEvent:
		if l.listener.OnEvent != nil {
			l.err = l.listener.OnEvent(p.data.(listener.EventData))
		}
	case packetTypeOnKVPair:
		if l.listener.OnKVPair != nil {
			l.err = l.listener.OnKVPair(p.data.(listener.KVPairData))
		}
	case packetTypeOnObjectUpdate:
		if l.listener.OnObjectUpdate != nil {
			l.err = l.listener.OnObjectUpdate(p.data.(listener.ObjectUpdateData))
		}
	case packetTypeCommit:
		if l.listener.Commit != nil {
			l.err = l.listener.Commit()
		}
		l.commitDoneChan <- l.err
	default:
		l.err = fmt.Errorf("unknown packet type: %d", p.packetType)
	}
	return false
}
