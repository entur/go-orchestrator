package orchestrator

import (
	"encoding/base64"
	"encoding/json"

	cloudevent "github.com/cloudevents/sdk-go/v2/event"
)

const DefaultMockRequestID = "mockid"                             // Default Request ID used in PO request mocks.
const DefaultMockContextID = "mockid"                             // Default Context ID used in PO request mocks.
const DefaultMockResponseTopic = "mocktopic"                      // Default Topic ID used in PO request mocks.
const DefaultMockPullRequestState = PullRequestStateOpen          // Default Pull Request state used in PO request mocks.
const DefaultMockRepositoryName = "mockrepo"                      // Default Repository name used in PO request mocks.
const DefaultMockRepositoryFullName = "entur/mockrepo"            // Default Repository full name used in PO request mocks.
const DefaultMockDefaultBranch = "main"                           // Default Repository branch used in PO request mocks.
const DefaultMockRepositoryVisibility = RepositoryVisbilityPublic // Default Repository visibility used in PO request mocks.
const DefaultMockSenderType = SenderTypeUser                      // Default Repository branch used in PO request mocks.
const DefaultMockUsername = "mockuser"                            // Default Github username used in PO request mocks.
const DefaultMockUserEmail = "mockuser@entur.io"                  // Default verified user email used in PO request mocks.
const DefaultMockUserPermission = RepositoryPermissionAdmin       // Default Repository permissions used in PO request mocks.
const DefaultMockAction = ActionPlan                              // Default User action used in PO request mocks.

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
		req.Resources.IAMLookup.URL = url
	}
}

func NewMockRequest(manifest any, opts ...MockRequestOption) (*Request, error) {
	newManifest, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	req := &Request{
		APIVersion: "orchestrator.entur.io/request/v1",
		Metadata: RequestMetadata{
			RequestID: DefaultMockRequestID,
			ContextID: DefaultMockContextID,
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
