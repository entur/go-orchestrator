# go-orchestrator

Go-Orchestrator is intended as a simple-to-use SDK for writing Sub-Orchestrators at Entur, as specified by the [Platform Orchestrator specification](https://github.com/entur/platform-orchestrator/tree/main/docs/architecture/reference/v1). 
It is written in Golang, and contains predefined type declarations, jsonschema validation rules, handler versioning, and mockers to make it as easy as possible to safely implement your dream Sub-Orchestrator.

## Quickstart

### Install 
```
go get github.com/entur/go-orchestrator
go mod tidy
```

### Import
```go
import (
	"github.com/entur/go-logging"
	"github.com/entur/go-orchestrator"
)
```

### Basic Usage
```go
import (
	"os"
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/entur/go-logging"
	"github.com/entur/go-orchestrator"
)


// -----------------------
// Initialize Cloud Function
// -----------------------

func init() {
	// Read Config!
	projectID := os.Getenv("PROJECT_ID")
	functionEntrypoint := os.Getenv("FUNCTION_ENTRYPOINT")

	// Setup Sub-Orchestrator!
	mh := NewMyMinimalManifestHandler()
	so := NewMyMinimalSubOrch(projectID, mh)

	// Start Cloud Function!
	h := orchestrator.NewCloudEventHandler(so)
	functions.CloudEvent(functionEntrypoint, h)
}

// -----------------------
// Sub-Orchestrator
// -----------------------

type MyMinimalSubOrch struct {
	projectID string
	handlers  []orchestrator.ManifestHandler
}

func (so *MyMinimalSubOrch) ProjectID() string {
	return so.projectID
}

func (so *MyMinimalSubOrch) Handlers() []orchestrator.ManifestHandler {
	return so.handlers
}

func NewMyMinimalSubOrch(projectID string, handlers ...orchestrator.ManifestHandler) *MyMinimalSubOrch {
	return &MyMinimalSubOrch{
		projectID: projectID,
		handlers:  handlers,
	}
}

// -----------------------
// Sub-Orchestrator Manifest Handlers
// -----------------------

// Your Manifest Definition --V
type MyMinimalManifest struct {
	orchestrator.ManifestHeader
	Metadata MyMinimalManifestMetadata `json:"metadata"`
	Spec     MyMinimalManifestSpec     `json:"spec"`
}

type MyMinimalManifestMetadata = orchestrator.ManifestMetadata // Default metadata definition, but you can use your own

type MyMinimalManifestSpec struct {
	Your   string   `json:"your"`
	Values []string `json:"values"`
	Here   int      `json:"here"`
}

// Your Manifest Handler ---V
type MyMinimalManifestHandler struct{}

func (h *MyMinimalManifestHandler) APIVersion() orchestrator.APIVersion {
	return "orchestrator.entur.io/MyMinimalSubOrch/v1" // Which Manifest version this handler operates on
}

func (h *MyMinimalManifestHandler) Kind() orchestrator.Kind {
	return "MyMinimalManifest" // Which Manifest Kind this handler operates on
}

func (h *MyMinimalManifestHandler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("plan not implemented")
}

func (h *MyMinimalManifestHandler) PlanDestroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("plan_destroy not implemented")
}

func (h *MyMinimalManifestHandler) Apply(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("apply not implemented")
}

func (h *MyMinimalManifestHandler) Destroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("destroy not implemented")
}

func NewMyMinimalManifestHandler() *MyMinimalManifestHandler {
	return &MyMinimalManifestHandler{}
}

```

## Overview and Architecture

### General Interfaces
The two most important types in this SDK are the `Orchestrator` and `ManifestHandler` interfaces.

An `Orchestrator` operates in a set google project, and has a set of `ManifestHandler`s that can work on manifests with a given pair of `apiVersion` and `kind`.
This enables your orchestrator to handle v1 and v2 of a given kind, and of course multiple differing kinds.

The `Orchestrator` interface is as follows:

```go
type Orchestrator interface {
	ProjectID() string           // The project this orchestrator is running in
	Handlers() []ManifestHandler // The manifests this orchestrator can handle
}
```

The `ManifestHandler` interface is as follows:

```go
type ManifestHandler interface {
	APIVersion() APIVersion // Which APIVersion this handler operates on
	Kind() Kind             // Which Kind this handler operates on
	// Actions
	Plan(context.Context, Request, *Result) error
	PlanDestroy(context.Context, Request, *Result) error
	Apply(context.Context, Request, *Result) error
	Destroy(context.Context, Request, *Result) error
}
```

### Middleware
It is possible to define middlewares that will run before or after Plan/Apply/PlanDestroy/Destroy actions. These can be defined at the Sub-Orchestrator level, or in the ManifestHandler:

The Middleware interface definitions are as follows:

```go
type MiddlewareBefore interface {
	MiddlewareBefore(context.Context, Request, *Result) error
}

type MiddlewareAfter interface {
	MiddlewareAfter(context.Context, Request, *Result) error
}
```


### Handling Internal Errors
During the processing of a Platform Orchestrator Request in a Sub-Orchestrator, unexpected internal errors might occur that should prevent any further processing from taking place. That being later in the function itself, or in a later middleware handler.
To handle such errors appropriately, the error should be returned immediately from the handler where it originated, such that the Go-Orchestrator SDK can log the event, and report an "An internal error occurred" to the user.

```go

func ErrorGeneratingFunction() error {
	return fmt.Errorf("this is an internal error")
}

func (h *Handler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	err := ErrorGeneratingFunction()
	if err != nil {
		return err // The internal error is not shown to the end-user, only logged. The user will instead see a generic "An internal error occurred" message
	}

	// We don't want to reach here

	return nil
}
```

Note: Returning an error will stop any further processing from occurring in later handlers.
Note2: User failures are not to be handled as internal errors, and should not return an error value!

### Handling User Errors
During the processing of a Platform Orchestrator Request in a Sub-Orchestrator, all unauthorized or invalid events (e.g. manifests containing invalid values) should result in a understandable failure message that is reported to the end-user.
To handle such failures approriately, the end result should be marked as having failed using the `r.Fail()` method with a informative message. Later processing steps should also be skipped by returning a nil value. Failing to return a nil value, will result in the error being handled as an internal error instead.

```go

func (h *Handler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest MyManifest
	
	err := json.Unmarshal(req.Manifest.New, &manifest)
	if err != nil {
		r.Fail(fmt.Sprintf("Invalid manifest:\n%s", err.Error())) // The message which the end-user sees in the PR comment.
		return nil
	}

	// We don't want to reach here if the manifest is invalid

	return nil
}
```

Note: Marking a result as having failed will stop later Plan/Apply/PlanDestroy/Destroy actions from processing, but not Middleware handlers. I.e. marking a `plan` action's result with `Fail` in a `MiddlewareBefore()` handler will stop the `Plan()` handler from running, but any `MiddleWareAfter()` handlers. You can check the current state of a result by calling `r.Locked()` and `r.Code()`

### Handling User Planned and Applied Changes
During the processing of a Platform Orchestrator Request in a Sub-Orchestrator, all planned and/or applied changes made in response the action and manifest contents should be grouped into create, update, or delete changesets. This can be done using the using the `r.Create()`, `r.Update()` and `r.Delete()` methods. Changes can either be represented as basic `string` types, or objects matching the Stringer/Change interface - i.e. types implementing a `String() string` method. The context of what a change might represent internally, will vary between Sub-Orchestrators. When all changes have been submitted, the final result should be marked as having succeeded using `r.Succeed()` with an informative summary. If a result is marked as having succeeded, but no changes have been added to the result, the final result code will be a `noop` as described in the [Platform Orchestrator specification](https://github.com/entur/platform-orchestrator/blob/main/docs/architecture/reference/v1/event-messages.md#result).

```go

type MyManifest struct {
	orchestrator.ManifestHeader
	Metadata MyMinimalManifestMetadata `json:"metadata"`
	Spec     MyMinimalManifestSpec     `json:"spec"`
}

type MyManifestManifestMetadata = orchestrator.ManifestMetadata // Default metadata definition, but you can use your own

type MyManifestManifestSpec struct {
	ClubName   string   `json:"clubName"`
}

type AdvancedChange struct {
	Plan TerraformPlan
}

func (change AdvancedChange) String() string {
	return fmt.Sprintf("%d terraform resources", change.Plan.NumberOfResources)
}

func (h *Handler) Apply(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest MyManifest
	
	err := json.Unmarshal(req.Manifest.New, &manifest)
	if err != nil {
		return nil
	}

	// Some terraform logic

	plan := Terraform.Plan()

	// End of terraform logic

	r.Create(AdvancedChange{
		Plan: plan,
	})

	r.Succeed(fmt.Sprintf("Applying the following changes to the GCP application '%s':", manifest.Metadata.ID)
	return nil
}
```

Note: If a result containing no changes (I.e.`r.Create()`, `r.Update()` and `r.Delete()` have not been called) is marked as having succeeded `r.Succeed()`, the final result code will be a `noop`.

## Run tests

This is a great way to get hacking! Simply modify the examples and play around.
Change the expected output followed by `// Output:` to verify your expectations.

```sh
go test ./...
go test orchestrator_minimal_example_test.go
go test orchestrator_example_test.go
```

## Examples
Interested in seeing how a Sub-Orchestrator might be implemented in practice? Take a look at the following examples and be inspired:
* See `./examples/minimal_suborchestrator` for a minimal sub-orchestrator implementation with tests.
* See `./examples/advanced_suborchestrator` for an advanced sub-orchestrator implementation with tests.
