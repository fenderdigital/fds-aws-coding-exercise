package main

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/fenderdigital/fds-aws-coding-exercise/src/handlers"
)

var (
	tableName string
	ddbCli    *ddb.Client
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch {
	case req.HTTPMethod == "GET" && strings.HasPrefix(req.Path, "/api/v1/subscriptions/"):
		userID := req.PathParameters["userId"]
		/*
			if userID == "" {
				parts := strings.Split(req.Path, "/")
				if len(parts) >= 5 {
					userID = parts[4]
				}
			}

		*/
		if userID == "" {
			return badRequest("missing userId"), nil
		}
		return handleGetSubscription(ctx, userID)
	default:
		return notFound("route not found"), nil
	}
}

func handleGetSubscription(ctx context.Context, userID string) (events.APIGatewayProxyResponse, error) {
	subs, err := handlers.GetUserSubs(ctx, ddbCli, tableName, userID)
	if err != nil {
		return serverErr(err), nil
	}

	return parseJSON(subs)
}

func main() {
	lambda.Start(handler)
}
