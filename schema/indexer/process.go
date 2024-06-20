package indexer

import (
	"cosmossdk.io/schema/listener"
)

type listenerProcess struct {
	listener       listener.Listener
	packetChan     chan listener.Packet
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

func (l *listenerProcess) processPacket(p listener.Packet) (stop bool) {
	if l.err != nil {
		if _, ok := p.(listener.Commit); ok {
			l.commitDoneChan <- l.err
			return true
		}
		return false
	}

	l.err = l.listener.ApplyPacket(p)
	if _, ok := p.(listener.Commit); ok {
		l.commitDoneChan <- l.err
		return l.err != nil
	}

	return false
}
