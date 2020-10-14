package types

import fmt "fmt"

const (
	// ModuleName defines the IBC convo name
	ModuleName = "convo"

	// Version defines the current version the IBC convo
	// module supports
	Version = "convo-v1"

	// PortID is the default port id that transfer module binds to
	PortID = "conversation"

	// StoreKey is the store key string for IBC transfer
	StoreKey = ModuleName

	// RouterKey is the message route for IBC transfer
	RouterKey = ModuleName

	// QuerierRoute is the querier route for IBC transfer
	QuerierRoute = ModuleName
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey = []byte{0x01}
)

// KeyPendingMessage returns the key under which the pending message that sender attempts to send
// over the sourceChannel to the receiver is stored
// If the message has been successfully received, then the pending message is removed
// and moved to the outbox
func KeyPendingMessage(sender, sourceChannel, receiver string) []byte {
	return []byte(fmt.Sprintf("%s/%s/%s/pending", sender, sourceChannel, receiver))
}

// KeyInboxMessage returns the key under which the last confirmed message sent from the sender to the
// receiver over the sourceChannel is stored
// For now, this key will only contain the latest message, all previous message sent by sender are erased.
func KeyInboxMessage(sender, sourceChannel, receiver string) []byte {
	return []byte(fmt.Sprintf("%s/%s/%s/inbox", sender, sourceChannel, receiver))
}

// KeyOutboxMessage returns the key under which the last received message by a sender to this receiver
// over the destination channel is stored
// For now, this key will only contain the latest message, all previously received messages will be erased.
func KeyOutboxMessage(receiver, destChannel, sender string) []byte {
	return []byte(fmt.Sprintf("%s/%s/%s/inbox", receiver, destChannel, sender))
}
