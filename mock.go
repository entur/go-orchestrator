package orchestrator

import (
	"encoding/json"
)

const MockRequestID = "mockid"
const MockResponseTopic = "mocktopic"
const MockDefaultBranch = "main"
const MockUserEmail = "mockuser@entur.io"

type MockRequestOption func(*Request)

func NewMockRequest(action Action, manifest any, opts ...MockRequestOption) (*Request, error) {
	newManifest, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	req := &Request{
		ApiVersion: "orchestrator.entur.io/request/v1",
		Metadata: OuterMetadata{
			RequestID: MockRequestID,
		},
		Origin: Origin{
			Repository: Repository{
				DefaultBranch: MockDefaultBranch,
				Visibility:    RepositoryVisbilityPublic,
			},
		},
		Sender: Sender{
			Type: SenderTypeUser,
			Email: MockUserEmail,
		},
		Action:        action,
		ResponseTopic: MockResponseTopic,
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
