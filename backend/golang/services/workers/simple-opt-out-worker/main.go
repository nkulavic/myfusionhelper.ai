package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"myfusionhelper.ai/internal/helpers"
	"myfusionhelper.ai/internal/helpers/automation"
)

func main() {
	// Register ONLY simple_opt_out helper
	helpers.Register("simple_opt_out", automation.NewSimpleOptOut)
	lambda.Start(Handle)
}

func Handle(ctx context.Context, event events.SQSEvent) error {
	// Worker logic here
	return nil
}
