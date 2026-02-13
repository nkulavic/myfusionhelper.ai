package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"myfusionhelper.ai/internal/helpers"
	"myfusionhelper.ai/internal/helpers/automation"
)

func main() {
	// Register ONLY goal_it helper
	helpers.Register("goal_it", automation.NewGoalIt)
	lambda.Start(Handle)
}

func Handle(ctx context.Context, event events.SQSEvent) error {
	// Worker logic here
	return nil
}
