package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"myfusionhelper.ai/internal/helpers"
	"myfusionhelper.ai/internal/helpers/integration"
)

func main() {
	// Register ONLY excel_it helper
	helpers.Register("excel_it", integration.NewExcelIt)
	lambda.Start(Handle)
}

func Handle(ctx context.Context, event events.SQSEvent) error {
	// Worker logic here
	return nil
}
