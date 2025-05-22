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

func ParseEvent(e event.Event) (Request, error) {
	var req Request
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

type MockEventOption func(*Request)

func NewMockEvent(manifest any, sender SenderType, action Action, options ...MockEventOption) (*event.Event, error) {

	newManifest, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	req := &Request{
		ApiVersion: "orchestrator.entur.io/request/v1",
		Metadata: OuterMetadata{
			RequestID: "mockid",
		},
		Sender: Sender{
			Type: sender,
		},
		Action:        action,
		ResponseTopic: "topic",
		Manifest:      Manifests{Old: nil, New: newManifest},
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
