package event

import (
	"encoding/base64"
	"encoding/json"

	cloudevent "github.com/cloudevents/sdk-go/v2/event"
	"github.com/entur/go-orchestrator"
)

type MockEventOption = orchestrator.MockRequestOption

func NewMockEvent(action orchestrator.Action, manifest any, opts ...MockEventOption) (*cloudevent.Event, error) {
	req, err := orchestrator.NewMockRequest(action, manifest, opts...)
	if err != nil {
		return nil, err
	}

	reqdata, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, base64.StdEncoding.EncodedLen(len(reqdata)))
	base64.StdEncoding.Encode(buf, reqdata)
	data, err := json.Marshal(&CloudEventData{
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

	e := cloudevent.New(cloudevent.CloudEventsVersionV03)
	e.DataEncoded = data
	return &e, nil
}
