package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"myfusionhelper.ai/internal/helpers"
	"myfusionhelper.ai/internal/helpers/contact"
)

func main() {
	// Register ONLY merge_it helper
	helpers.Register("merge_it", contact.NewMergeIt)
	lambda.Start(Handle)
}

func Handle(ctx context.Context, event events.SQSEvent) error {
	// Worker logic here
	return nil
}
