package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/fenderdigital/fds-aws-coding-exercise/src/dtos"
	"github.com/fenderdigital/fds-aws-coding-exercise/src/handlers"
)

type LambdaHandler struct {
	apiHandler *handlers.ApiHandler
}

func NewLambdaHandler(ctx context.Context) (*LambdaHandler, error) {
	tableName := os.Getenv("DDB_TABLE")
	if tableName == "" {
		return nil, fmt.Errorf("DDB_TABLE is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return &LambdaHandler{apiHandler: handlers.NewApiHandler(tableName, ddb.NewFromConfig(cfg))}, nil
}

func (lh *LambdaHandler) handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch {
	case req.HTTPMethod == "GET" && strings.HasPrefix(req.Path, "/api/v1/subscriptions/"):
		userID := req.PathParameters["userId"]
		if userID == "" {
			return badRequest("missing userId"), nil
		}
		return lh.handleGetSubscription(ctx, userID)
	case req.HTTPMethod == "POST" && req.Path == "/api/v1/webhooks/subscriptions":
		var subEventReq dtos.SubscriptionRequest
		if err := json.Unmarshal([]byte(req.Body), &subEventReq); err != nil {
			return badRequest(err.Error()), nil
		}
		switch subEventReq.EventType {
		case "subscription.created":
			return lh.handleCreateSubscription(ctx, &subEventReq)
		case "subscription.renewed":
			return lh.handleRenewSubscription(ctx, &subEventReq)
		case "subscription.cancelled":
			return lh.handleCancelSubscription(ctx, &subEventReq)
		default:
			return badRequest("unknown event type"), nil
		}
	}

	return notFound("route not found"), nil
}

func (lh *LambdaHandler) handleGetSubscription(ctx context.Context, userID string) (events.APIGatewayProxyResponse, error) {
	subs, err := lh.apiHandler.GetUserSubs(ctx, userID)
	if err != nil {
		return serverErr(err), nil
	}

	return parseJSON(http.StatusOK, subs)
}

func (lh *LambdaHandler) handleCreateSubscription(ctx context.Context, req *dtos.SubscriptionRequest) (events.APIGatewayProxyResponse, error) {
	err := lh.apiHandler.CreateUserSub(ctx, req)
	if err != nil {
		return serverErr(err), nil
	}
	return parseJSON(http.StatusCreated, map[string]string{"status": "created"})
}

func (lh *LambdaHandler) handleRenewSubscription(ctx context.Context, req *dtos.SubscriptionRequest) (events.APIGatewayProxyResponse, error) {
	err := lh.apiHandler.RenewUserSub(ctx, req)
	if err != nil {
		return serverErr(err), nil
	}

	return parseJSON(http.StatusOK, map[string]string{"status": "renewed"})
}

func (lh *LambdaHandler) handleCancelSubscription(ctx context.Context, req *dtos.SubscriptionRequest) (events.APIGatewayProxyResponse, error) {
	err := lh.apiHandler.CancelUserSub(ctx, req)
	if err != nil {
		return serverErr(err), nil
	}
	return parseJSON(http.StatusOK, map[string]string{"status": "cancelled"})
}

func main() {
	ctx := context.Background()
	lHandler, err := NewLambdaHandler(ctx)
	if err != nil {
		log.Fatal(err)
	}
	lambda.Start(lHandler.handler)
}

func parseJSON(code int, v any) (events.APIGatewayProxyResponse, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: code,
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

func requestErr(err error) events.APIGatewayProxyResponse {
	// Maybe not the best error handling for now, update later to use errors.Is/errors.As if possible :)
	if strings.Contains(err.Error(), "active/pending") {
		return events.APIGatewayProxyResponse{StatusCode: 409, Body: fmt.Sprintf(`{"error":"%s"}`, err.Error())}
	}
	return events.APIGatewayProxyResponse{StatusCode: 422, Body: fmt.Sprintf(`{"error":"%s"}`, err.Error())}
}
