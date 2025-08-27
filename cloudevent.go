package orchestrator

import (
	"encoding/json"

	cloudevent "github.com/cloudevents/sdk-go/v2/event"
)

// -----------------------
// Cloud Event
// -----------------------

type PubSubMessageAttributes struct{}

type PubSubMessage struct {
	ID          string                  `json:"messageId"`
	PublishTime string                  `json:"publishTime"`
	Attributes  PubSubMessageAttributes `json:"attributes"`
	Data        []byte                  `json:"data"`
}

type CloudEventData struct {
	Subscription string
	Message      PubSubMessage
}

func UnmarshalCloudEvent(e cloudevent.Event, v any) error {
	var data CloudEventData
	err := e.DataAs(&data)
	if err != nil {
		return err
	}

	return json.Unmarshal(data.Message.Data, v)
}