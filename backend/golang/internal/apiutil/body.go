package apiutil

import (
	"encoding/base64"

	"github.com/aws/aws-lambda-go/events"
)

// GetBody returns the decoded request body from an API Gateway v2 event.
// API Gateway HTTP API may base64-encode request bodies; this function
// transparently decodes when IsBase64Encoded is true.
func GetBody(event events.APIGatewayV2HTTPRequest) string {
	if event.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(event.Body)
		if err != nil {
			return event.Body
		}
		return string(decoded)
	}
	return event.Body
}
