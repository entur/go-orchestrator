package orchestrator

import (
	"encoding/json"

	"github.com/cloudevents/sdk-go/v2/event"
)

func ParseEvent[T any](e event.Event) (Request[T], error) {
	var req Request[T]
	var data CloudEventData
	err := e.DataAs(&data)
	if err != nil {
		return req, err
	}

	err = json.Unmarshal(data.Message.Data, &req)
	if err != nil {
		return req, err
	}
	return req, nil
}
func ParsePayload() {}
