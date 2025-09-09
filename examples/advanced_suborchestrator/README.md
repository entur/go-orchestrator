# Advanced Sub-Orchestrator

This example shows how to create an advanced sub-orchestrator, taking advantage of most of the Go-Orchestrator SDK features. It handles two different manfiest kinds and one version, here's the spec:

```yaml
apiVersion: orchestrator.entur.io/vehicle/v1
kind: Airplane
metadata:
  id: someid
spec:
  model: "Boeing 747"
  wingspanMeters: 45.6
  numberOfPassengers: 30
```

```yaml
apiVersion: orchestrator.entur.io/vehicle/v1
kind: Car
metadata:
  id: someid
spec:
  model: "Ford Fiesta"
  numberOfWheels: 4
  numberOfPassengers: 3
```

## Note

This example is not yet complete, and is yet to be expanded upon in the future.