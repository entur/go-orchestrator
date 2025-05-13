package orchestrator

import (
	"context"
	"encoding/json"

	"fmt"

	"cloud.google.com/go/pubsub"
)

func Process[T any](ctx context.Context, o Orchestrator[T], topic *pubsub.Topic, req Request[T]) error {
	var result Result
	var err error

	switch req.Action {
	case Apply:
		result, err = o.Apply(ctx, req)
	case Plan:
		result, err = o.Plan(ctx, req)
	case PlanDestroy:
		result, err = o.PlanDestroy(ctx, req)
	case Destroy:
		result, err = o.Destroy(ctx, req)
	}
	if err != nil {
		// TODO: create a Result that is an Error
		return err
	}

	response := req.ToResponse(result)

	err = Respond(ctx, topic, response)
	return err
}

func Respond(ctx context.Context, topic *pubsub.Topic, r Response) error {
	// logger := zerolog.Ctx(ctx)
	if topic == nil {
		return fmt.Errorf("no topic set, cannot respond")
	}
	enc, err := json.Marshal(r)
	if err != nil {
		return err
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: enc,
	})
	_, err = result.Get(ctx)
	return err
}
