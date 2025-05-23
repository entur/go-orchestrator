package events

import (
	"encoding/base64"
	"encoding/json"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/entur/go-orchestrator"
)

const MockUserEmail = "mockuser@entur.io"

type MockEventOption func(*orchestrator.Request)

func NewMockEvent(manifest any, sender orchestrator.SenderType, action orchestrator.Action, options ...MockEventOption) (*event.Event, error) {
	newManifest, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	req := &orchestrator.Request{
		ApiVersion: "orchestrator.entur.io/request/v1",
		Metadata: orchestrator.OuterMetadata{
			RequestID: "mockid",
		},
		Sender: orchestrator.Sender{
			Type: sender,
		},
		Action:        action,
		ResponseTopic: "topic",
		Manifest:      orchestrator.Manifests{Old: nil, New: newManifest},
	}
	if sender == orchestrator.SenderTypeUser {
		req.Sender.Email = MockUserEmail
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
