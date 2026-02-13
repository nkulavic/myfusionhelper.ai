package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"myfusionhelper.ai/internal/helpers"
	"myfusionhelper.ai/internal/helpers/analytics"
)

func main() {
	// Register ONLY customer_lifetime_value helper
	helpers.Register("customer_lifetime_value", analytics.NewCustomerLifetimeValue)
	lambda.Start(Handle)
}

func Handle(ctx context.Context, event events.SQSEvent) error {
	// Worker logic here
	return nil
}
