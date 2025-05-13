package orchestrator

import (
	"context"
	"encoding/json"

	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog"
)

func Process[T any](ctx context.Context, o Orchestrator[T], topic *pubsub.Topic, req Request[T]) error {
	logger := zerolog.Ctx(ctx)
	var result Result
	var err error

	switch req.Action {
	case ActionApply:
		result, err = o.Apply(ctx, req)
	case ActionPlan:
		result, err = o.Plan(ctx, req)
	case ActionPlanDestroy:
		result, err = o.PlanDestroy(ctx, req)
	case ActionDestroy:
		result, err = o.Destroy(ctx, req)
	}
	if err != nil {
		result.Code = resultCodeError
		result.Summary = err.Error()
		result.Creations = nil
		result.Updates = nil
		result.Deletions = nil
	}

	response := req.ToResponse(result)
	logger.Info().Msg(response.Output)
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
