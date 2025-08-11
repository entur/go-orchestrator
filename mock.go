package orchestrator

import (
	"encoding/json"
)

const MockUserEmail = "mockuser@entur.io"

type MockRequestOption func(*Request)

func NewMockRequest(manifest any, sender SenderType, action Action, opts ...MockRequestOption) (*Request, error) {
	newManifest, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	req := &Request{
		ApiVersion: "orchestrator.entur.io/request/v1",
		Metadata: OuterMetadata{
			RequestID: "mockid",
		},
		Origin: Origin{
			Repository: GitRepository{
				DefaultBranch: "main",
				Visibility:    GitRepositoryVisbilityPublic,
			},
		},
		Sender: Sender{
			Type: sender,
		},
		Action:        action,
		ResponseTopic: "topic",
		Manifest: Manifests{
			Old: nil,
			New: newManifest,
		},
	}
	if sender == SenderTypeUser {
		req.Sender.Email = MockUserEmail
	}

	for _, opt := range opts {
		opt(req)
	}

	return req, err
}
