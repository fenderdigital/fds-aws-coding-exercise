package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

func parseJSON(v any) (events.APIGatewayProxyResponse, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}, nil
}

func badRequest(msg string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{StatusCode: 400, Body: fmt.Sprintf(`{"error":"%s"}`, msg)}
}

func serverErr(err error) events.APIGatewayProxyResponse {
	// Not configuring log libraries for now
	log.Printf("server error: %v", err)
	return events.APIGatewayProxyResponse{StatusCode: 500, Body: `{"error":"internal error"}`}
}

func notFound(msg string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{StatusCode: 404, Body: fmt.Sprintf(`{"error":"%s"}`, msg)}
}
