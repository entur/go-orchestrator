# sub-orchestrator-sdk

> [!WARNING]  
> Work in progress!

## Start using the SDK

You need to enable private go modules from entur:

```sh
go env -w GOPRIVATE='github.com/entur/*'
```

## Minimal example

Here's a `main.go` to get you started.

```go
package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	orchestrator "github.com/entur/sub-orchestrator-sdk"
)

type ExampleManifest struct {
	ApiVersion orchestrator.ApiVersion `json:"apiVersion"`
	Kind       orchestrator.Kind       `json:"kind"`
	Metadata   orchestrator.Metadata   `json:"metadata"`
}

type ExampleSO struct {
	/* you can have some internal state here */
}

func (s *ExampleSO) Plan(ctx context.Context, req orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("plan not implemented")
}
func (s *ExampleSO) PlanDestroy(ctx context.Context, req orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("plandestroy not implemented")
}
func (s *ExampleSO) Apply(ctx context.Context, req orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("apply not implemented")
}
func (s *ExampleSO) Destroy(context.Context, orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("destroy not implemented")
}

func main() {
	// Just an example manifest, here is where you specify _your_ sub-orchestrator
	// ApiVersion, Kind and Metadata.Id is reqired
	manifest := ExampleManifest{
		ApiVersion: "orcestrator.entur.io/example/v1",
		Kind:       "Example",
		Metadata: orchestrator.Metadata{
			Id: "mything",
		},
	}

	so := ExampleSO{}
	project := "not-a-project" // os.Getenv("PROJECT_ID")
	client, _ := pubsub.NewClient(context.Background(), project)
	handler := orchestrator.NewEventHandler(&so, client)

	event, _ := orchestrator.NewMockCloudEvent(manifest, "plan")
	err := handler(context.Background(), *event)

	if err != nil {
		fmt.Println("HANDLER ERR:", err)
	}
}
```
