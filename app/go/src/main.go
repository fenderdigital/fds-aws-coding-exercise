package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

type Response struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func handler(ctx context.Context, event interface{}) (Response, error) {
	return Response{
		StatusCode: 200,
		Body:       "Hello from Go Lambda!",
	}, nil
}

func main() {
	lambda.Start(handler)
}
