# Minimal Sub-Orchestrator

This example shows how to create a minimal sub-orchestrator, taking advantage of the basic go-orchestrator SDK features. It handles one manfiest kind and version, here's the spec:

```yaml
apiVersion: orchestrator.entur.io/example/v1
kind: Example
metadata:
  id: someid
spec:
  name: Some name
```