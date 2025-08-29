package orchestrator

import (
	"encoding/base64"
	"encoding/json"

	cloudevent "github.com/cloudevents/sdk-go/v2/event"
)

const DefaultMockRequestID = "mockid"
const DefaultMockResponseTopic = "mocktopic"
const DefaultMockPullRequestState = PullRequestStateOpen
const DefaultMockRepositoryName = "mockrepo"
const DefaultMockRepositoryFullName = "entur/mockrepo"
const DefaultMockDefaultBranch = "main"
const DefaultMockRepositoryVisibility = RepositoryVisbilityPublic
const DefaultMockSenderType = SenderTypeUser
const DefaultMockUsername = "mockuser"
const DefaultMockUserEmail = "mockuser@entur.io"
const DefaultMockUserPermission = RepositoryPermissionAdmin
const DefaultMockAction = ActionPlan

type MockRequestOption func(*Request)

func WithAction(action Action) MockRequestOption {
	return func(req *Request) {
		req.Action = action
	}
}

func WithSender(sender Sender) MockRequestOption {
	return func(req *Request) {
		req.Sender = sender
	}
}

func WithIAMEndpoint(url string) MockRequestOption {
	return func(req *Request) {
		req.Resources.IAM.Url = url
	}
}

func NewMockRequest(manifest any, opts ...MockRequestOption) (*Request, error) {
	newManifest, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	req := &Request{
		ApiVersion: "orchestrator.entur.io/request/v1",
		Metadata: OuterMetadata{
			RequestID: DefaultMockRequestID,
		},
		Origin: Origin{
			PullRequest: PullRequest{
				State: DefaultMockPullRequestState,
			},
			Repository: Repository{
				Name:          DefaultMockRepositoryName,
				FullName:      DefaultMockRepositoryFullName,
				DefaultBranch: DefaultMockDefaultBranch,
				Visibility:    DefaultMockRepositoryVisibility,
			},
		},
		Sender: Sender{
			Username:   DefaultMockUsername,
			Email:      DefaultMockUserEmail,
			Type:       DefaultMockSenderType,
			Permission: DefaultMockUserPermission,
		},
		Action:        DefaultMockAction,
		ResponseTopic: DefaultMockResponseTopic,
		Manifest: Manifests{
			Old: nil,
			New: newManifest,
		},
	}

	for _, opt := range opts {
		opt(req)
	}

	return req, err
}

func NewMockCloudEvent(manifest any, opts ...MockRequestOption) (*cloudevent.Event, error) {
	req, err := NewMockRequest(manifest, opts...)
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
