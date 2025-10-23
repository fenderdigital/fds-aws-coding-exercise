package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fenderdigital/fds-aws-coding-exercise/src/dtos"
	"github.com/fenderdigital/fds-aws-coding-exercise/src/entities"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func GetUserSubs(ctx context.Context, ddbCli *ddb.Client, tableName, userID string) ([]dtos.SubscriptionResponse, error) {
	pk := "user:" + userID
	out, err := ddbCli.Query(ctx, &ddb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("#pk = :pk"),
		ExpressionAttributeNames: map[string]string{
			"#pk": "pk",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query user subscription: %w", err)
	}

	subs := make([]entities.SubscriptionItem, 0, len(out.Items))
	for _, it := range out.Items {
		var s entities.SubscriptionItem
		if err := attributevalue.UnmarshalMap(it, &s); err != nil {
			return nil, fmt.Errorf("failed to unmarshal subscription item: %w", err)
		}
		subs = append(subs, s)
	}

	subResponses := make([]dtos.SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		plan, err := getPlan(ctx, ddbCli, tableName, sub.PlanSKU)
		if err != nil {
			return nil, err
		}

		subID := strings.TrimPrefix(sub.SK, "sub:")

		status, err := getStatus(sub.CanceledAt, sub.ExpiresAt)
		if err != nil {
			return nil, err
		}

		resp := dtos.SubscriptionResponse{
			UserID:         strings.TrimPrefix(sub.PK, "user:"),
			SubscriptionID: subID,
			StartDate:      sub.StartDate,
			ExpiresAt:      sub.ExpiresAt,
			CancelledAt:    sub.CanceledAt,
			Status:         status,
			Attributes:     sub.Attributes,
			Plan: &dtos.SubscriptionResponsePlan{
				SKU:          plan.SKU,
				Name:         plan.Name,
				Price:        plan.Price,
				Currency:     plan.Currency,
				BillingCycle: plan.BillingCycle,
				Features:     plan.Features,
			},
		}

		subResponses = append(subResponses, resp)
	}

	return subResponses, nil
}

func getPlan(ctx context.Context, ddbCli *ddb.Client, tableName, sku string) (*entities.Plan, error) {
	pk, sk := planKey(sku)
	out, err := ddbCli.GetItem(ctx, &ddb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, fmt.Errorf("plan %s not found", sku)
	}
	var p entities.Plan
	if err := attributevalue.UnmarshalMap(out.Item, &p); err != nil {
		return nil, err
	}

	if strings.HasPrefix(p.PK, "plan:") {
		p.SKU = strings.TrimPrefix(p.PK, "plan:")
	}
	return &p, nil
}

func getStatus(canceledAt *string, expiresAt string) (dtos.SubStatus, error) {
	exp, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return "", fmt.Errorf("invalid expiresAt parsing: %w", err)
	}
	if canceledAt == nil || *canceledAt == "" {
		return dtos.SubStatusActive, nil
	}

	if time.Now().Before(exp) {
		return dtos.SubStatusPending, nil
	}

	return dtos.SubStatusCancelled, nil
}

func planKey(sku string) (pk, sk string) {
	return "plan:" + sku, "meta"
}
