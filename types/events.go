package types

import "fmt"

// EventManager is a naive tag collector that represents a series of "events"
// that modules can emit. Its sole purpose is to allow for duplicate tags by
// key until a more complete solution exists via an upstream Tendermint upgrade.
type EventManager struct {
	events map[string]Tags
}

// NewEventManager returns a reference to a new EventManager.
func NewEventManager() *EventManager {
	return &EventManager{
		events: make(map[string]Tags),
	}
}

// Event collects a series of tags.
func (em *EventManager) Event(tags Tags) {
	for _, tag := range tags {
		key := string(tag.Key)

		if _, ok := em.events[key]; ok {
			em.events[key] = append(em.events[key], tag)
		} else {
			em.events[key] = Tags{tag}
		}
	}
}

// ToTags flattens all collected/emitted events where for any duplicate keys, each
// key will contain a ordinal suffix (e.g. action-3).
func (em EventManager) ToTags() Tags {
	var resTags Tags

	for _, tags := range em.events {
		if len(tags) > 1 {
			for i, tag := range tags {
				resTags = resTags.AppendTag(fmt.Sprintf("%s-%d", string(tag.Key), i+1), tag.Value)
			}
		} else {
			resTags = append(resTags, tags[0])
		}
	}

	return resTags
}
