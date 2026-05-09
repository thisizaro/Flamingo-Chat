package events

import "encoding/json"

func BuildEvent(
	eventType EventType,
	payload any,
) ([]byte, error) {

	event := Event{
		Type: eventType,
		Payload: payload,
	}

	return json.Marshal(event)
}