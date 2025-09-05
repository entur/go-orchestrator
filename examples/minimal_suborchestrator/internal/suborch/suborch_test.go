package suborch

import (
	"context"
	"errors"
	"testing"

	"github.com/entur/go-orchestrator"
)

// -----------------------
// Sub-Orchestrator Integration Test
// -----------------------
func TestIntegration(t *testing.T) {
	mh := NewMyMinimalManifestHandler()
	so := NewMyMinimalSubOrch("mock", mh)

	// Our mock manifest that we want to test
	mockManifest := MyMinimalManifest{
		ManifestHeader: orchestrator.ManifestHeader{
			APIVersion: "orchestrator.entur.io/MyMinimalSubOrch/v1",
			Kind:       "MyMinimalManifest",
		},
		Metadata: MyMinimalManifestMetadata{
			ID: "manifest-id",
		},
		Spec: MyMinimalManifestSpec{
			Your:   "Some value",
			Values: []string{"One", "Two"},
			Here:   7,
		},
	}

	// Testing of full Cloud Event flow
	{
		handler := orchestrator.NewCloudEventHandler(so, orchestrator.WithCustomPubSubClient(nil))
		mockEvent, _ := orchestrator.NewMockCloudEvent(mockManifest)
		err := handler(context.Background(), *mockEvent)
		if err != nil {
			t.Errorf("cloud event handler returned non-nil error \ngot: %s", err.Error())
		}
	}

	// Testing of Sub-Orchestrator processing logic only
	{
		mockRequest, _ := orchestrator.NewMockRequest(mockManifest)
		result := orchestrator.Process(context.Background(), so, mockRequest)

		errs := result.Errors()

		if len(errs) > 0 {
			err := errors.Join(errs...)
			t.Logf("result had some errors:\n%s", err.Error())
		}

		code := result.Code()
		creations := result.Creations()
		updates := result.Updates()
		deletions := result.Deletions()

		if code != orchestrator.ResultCodeSuccess {
			t.Errorf("result status code did not match the expected code\ngot: %s\nwant: %s", code, orchestrator.ResultCodeSuccess)
		}

		num := len(creations)
		if num != 1 {
			t.Errorf("result creations did not match the number of expected changes\ngot: %d\nwant: %d", num, 1)
		}
		num = len(updates)
		if num != 1 {
			t.Errorf("result updates did not match the number of expected changes\ngot: %d\nwant: %d", num, 1)
		}
		num = len(deletions)
		if num != 0 {
			t.Errorf("result deletions did not match the number of expected changes\ngot: %d\nwant: %d", num, 0)
		}
	}
}
