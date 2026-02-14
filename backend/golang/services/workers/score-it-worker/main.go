package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/myfusionhelper/api/internal/helpers"
	"github.com/myfusionhelper/api/internal/helpers/tagging"
	"github.com/myfusionhelper/api/internal/worker"

	// Register all connectors via init()
	_ "github.com/myfusionhelper/api/internal/connectors"
)

func main() {
	helpers.Register("score_it", tagging.NewScoreIt)
	lambda.Start(worker.HandleSQSEvent)
}
