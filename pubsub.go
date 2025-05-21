package orchestrator

import (
	"encoding/base64"
	"encoding/json"

	"github.com/cloudevents/sdk-go/v2/event"
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

func ParseEvent[T any](e event.Event) (Request[T], error) {
	var req Request[T]
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

type MockEventOption[T any] func(*Request[T])

func NewMockEvent[T any](manifest T, sender SenderType, action Action, options ...MockEventOption[T]) (*event.Event, error) {
	req := &Request[T]{
		ApiVersion: "orchestrator.entur.io/request/v1",
		Metadata: OuterMetadata{
			RequestID: "mockid",
		},
		Sender: Sender{
			Type: sender,
		},
		Action:        action,
		ResponseTopic: "topic",
		Manifest:      Manifests[T]{Old: nil, New: manifest},
	}
	for _, opt := range options {
		opt(req)
	}
	reqdata, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(reqdata)))
	base64.StdEncoding.Encode(buf, reqdata)
	data, err := json.Marshal(&EventData{
		Message: PubSubMessage{
			Data:        reqdata,
			ID:          "id",
			PublishTime: "time",
			Attributes:  PubSubMessageAttributes{},
		},
		Subscription: "sub",
	})
	if err != nil {
		return nil, err
	}

	event := event.New(event.CloudEventsVersionV03)
	event.DataEncoded = data
	return &event, nil
}
