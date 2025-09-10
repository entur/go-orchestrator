# Minimal Sub-Orchestrator

This example shows how to create a minimal sub-orchestrator, taking advantage of the basic Go-Orchestrator SDK features. It handles one manfiest kind and version, here's the spec:

```yaml
apiVersion: orchestrator.entur.io/MyMinimalSubOrch/v1
kind: MyMinimalManifest
metadata:
  id: someid
spec:
  your: "your"
  values:
    - "value"
  here: 1
```