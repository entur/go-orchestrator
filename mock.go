package orchestrator

import (
	"encoding/json"
)

const DefaultMockRequestID = "mockid"
const DefaultMockResponseTopic = "mocktopic"
const DefaultMockDefaultBranch = "main"
const DefaultMockRepositoryVisibility = RepositoryVisbilityPublic
const DefaultMockSenderType = SenderTypeUser
const DefaultMockUserEmail = "mockuser@entur.io"
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
			Repository: Repository{
				DefaultBranch: DefaultMockDefaultBranch,
				Visibility:    DefaultMockRepositoryVisibility,
			},
		},
		Sender: Sender{
			Type: DefaultMockSenderType,
			Email: DefaultMockUserEmail,
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
