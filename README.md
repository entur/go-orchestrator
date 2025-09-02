# go-orchestrator

## Start using the SDK

### High level overview and architecture

The two most important interfaces in this SDK is the `Orchestrator` and `ManifestHandler`.

An `Orchestrator` operates in a google project and has a set of `ManifestHandler`s that can work on manifests with a given pair of `apiVersion` and `kind`.
This enables your orchestrator to handle v1 and v2 of a given kind, and of course multiple differing kinds.

First, you will need a type that is an `Orchestrator`:

```go
type Orchestrator interface {
	ProjectID() string           // The project this orchestrator is running in
	Handlers() []ManifestHandler // The manifests this orchestrator can handle
}
```

Then you will need at least one `ManifestHandler`:

```go
type ManifestHandler interface {

	APIVersion() APIVersion // Which ApiVersion this handler operates on
	Kind() Kind             // Which Kind this handler operates on
	// Actions
	Plan(context.Context, Request, *Result) error
	PlanDestroy(context.Context, Request, *Result) error
	Apply(context.Context, Request, *Result) error
	Destroy(context.Context, Request, *Result) error
}
```

For further detail, please see one of the test files described below.

### Code setup

Typical dependencies

```go
import (
	"github.com/entur/go-logging"
	"github.com/entur/go-orchestrator"
)
```

Now

```sh
go mod tidy
```

## Run tests and examples

This is a great way to get hacking! Simply modify the examples and play around.
Change the expected output followed by `// Output:` to verify your expectations.

```sh
go test ./...
go test orchestrator_minimal_example_test.go
go test orchestrator_example_test.go
```

### Minimal example

This example creates a minimal orchestrator, using mostly default behavior. It handles one kind and one version, here's the spec:

```yaml
apiVersion: orchestrator.entur.io/example/v1
kind: Example
metadata:
  id: someid
spec:
  name: Some name
```

See `./orchestrator_minimal_example_test.go` for a minimal test and implementation.

### Full example

This example creates an orchestrator using most of the APIs in this SDK.

- Versioned types
- MiddlewareBefore (auth)
- MiddlewareAfter (audit log)
- Custom console log writer for testing
- Mock IAM server for testing

The code is written in a way to make it clear that a future v2 may come and serves as a best practice reference.

```yaml
apiVersion: orchestrator.entur.io/example/v1
kind: Example
metadata:
  id: someid
spec:
  name: Some name
```

See `./orchestrator_example_test.go` for a complete test and implementation.
