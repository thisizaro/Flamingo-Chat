package events

import "encoding/json"

func ParseEvent(data []byte) (*Event, error) {

	var event Event

	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}