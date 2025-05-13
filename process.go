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

	var code ResultCode
	var msg string

	if err != nil {
		msg = "An internal error occured"
		code = ResultCodeError
		logger.Error().Err(err).Interface("result", result).Msg(msg)
	} else {
		msg = result.String()

		if result.Success == false {
			code = ResultCodeFailure
		} else if len(result.Creations) == 0 && len(result.Updates) == 0 && len(result.Deletions) == 0 {
			code = ResultCodeNoop
		} else {
			code = ResultCodeSuccess
		}
	}

	response := req.ToResponse(code, msg)
	logger.Info().Interface("response", response).Msg("Response ready to send")
	err = Respond(ctx, topic, response)
	if err != nil {
		logger.Error().Err(err).Msg("Could not respond")
	}
	
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
