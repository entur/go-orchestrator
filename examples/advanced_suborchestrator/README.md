# Advanced Sub-Orchestrator

This example shows how to create a minimal sub-orchestrator, taking advantage of the basic go-orchestrator SDK features. It handles one manfiest kind and version, here's the spec:

```yaml
apiVersion: orchestrator.entur.io/example/v1
kind: Example
metadata:
  id: someid
spec:
  name: Some name
```

This example creates an advanced sub-orchestrator using most of the APIs in this SDK.

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