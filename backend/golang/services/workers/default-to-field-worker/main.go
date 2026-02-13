package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"myfusionhelper.ai/internal/helpers"
	"myfusionhelper.ai/internal/helpers/contact"
)

func main() {
	// Register ONLY default_to_field helper
	helpers.Register("default_to_field", contact.NewDefaultToField)
	lambda.Start(Handle)
}

func Handle(ctx context.Context, event events.SQSEvent) error {
	// Worker logic here
	return nil
}
