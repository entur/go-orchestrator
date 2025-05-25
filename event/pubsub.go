package event

import (
	"encoding/json"

	cloudevent "github.com/cloudevents/sdk-go/v2/event"
	"github.com/entur/go-orchestrator"
)

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

func ParseEvent(e cloudevent.Event) (*orchestrator.Request, error) {
	var data CloudEventData
	err := e.DataAs(&data)
	if err != nil {
		return nil, err
	}

	var req orchestrator.Request
	err = json.Unmarshal(data.Message.Data, &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}
