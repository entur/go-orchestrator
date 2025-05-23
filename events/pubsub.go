package events

import (
	"encoding/json"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/entur/go-orchestrator"
)

type PubSubMessageAttributes struct{}

type PubSubMessage struct {
	ID          string                  `json:"messageId"`
	PublishTime string                  `json:"publishTime"`
	Attributes  PubSubMessageAttributes `json:"attributes"`
	Data        []byte                  `json:"data"`
}

type EventData struct {
	Subscription string
	Message      PubSubMessage
}

func ParseEvent(e event.Event) (orchestrator.Request, error) {
	var req orchestrator.Request
	var data EventData
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
