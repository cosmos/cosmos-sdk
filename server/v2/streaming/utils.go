package streaming

import "cosmossdk.io/core/event"

func IntoStreamingEvents(events []event.Event) ([]*Event, error) {
	streamingEvents := make([]*Event, len(events))

	for _, event := range events {
		strEvent := &Event{
			Type: event.Type,
		}
		attrs, err := event.Attributes()
		if err != nil {
			return nil, err
		}
		for _, eventValue := range attrs {
			strEvent.Attributes = append(strEvent.Attributes, &EventAttribute{
				Key:   eventValue.Key,
				Value: eventValue.Value,
			})
		}
		streamingEvents = append(streamingEvents, strEvent)
	}

	return streamingEvents, nil
}
