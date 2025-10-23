package handlers

import (
	"context"
	"errors"
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

var (
	errActivePlan         = errors.New("plan is active")
	errActiveOrPendingSub = errors.New("user already has an active/pending subscription")
	errMissingCanceledAt  = errors.New("missing canceledAt time to cancel subscription")
)

func GetUserSubs(ctx context.Context, ddbCli *ddb.Client, tableName, userID string) ([]*dtos.SubscriptionResponse, error) {
	subs, err := getUserSubs(ctx, ddbCli, tableName, userID)
	if err != nil {
		return nil, err
	}

	subResponses := make([]*dtos.SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		plan, err := getPlan(ctx, ddbCli, tableName, sub.PlanSKU)
		if err != nil {
			return nil, err
		}

		subID := strings.TrimPrefix(sub.SK, "sub:")

		status, err := getStatus(sub.CancelledAt, sub.ExpiresAt)
		if err != nil {
			return nil, err
		}

		resp := dtos.SubscriptionResponse{
			UserID:         strings.TrimPrefix(sub.PK, "user:"),
			SubscriptionID: subID,
			StartDate:      sub.StartDate,
			ExpiresAt:      sub.ExpiresAt,
			CancelledAt:    sub.CancelledAt,
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

		subResponses = append(subResponses, &resp)
	}

	return subResponses, nil
}

func CreateUserSub(ctx context.Context, ddbCli *ddb.Client, tableName string, subReq *dtos.SubscriptionRequest) error {
	plan, err := getPlan(ctx, ddbCli, tableName, asString(subReq.Metadata["planSku"]))
	if err != nil {
		return err
	}

	if strings.ToLower(plan.Status) != "active" {
		return errActivePlan
	}

	subs, err := getUserSubs(ctx, ddbCli, tableName, subReq.UserID)
	if err != nil {
		return err
	}

	hasActiveOrPendingSub, err := hasActiveOrPending(subs)
	if err != nil {
		return err
	}

	if hasActiveOrPendingSub {
		return errActiveOrPendingSub
	}

	pk, sk := userSubKey(subReq.UserID, subReq.SubscriptionID)
	start := subReq.Timestamp
	planSKU := asString(subReq.Metadata["planSku"])

	delete(subReq.Metadata, "planSku")
	item := entities.SubscriptionItem{
		PK:             pk,
		SK:             sk,
		Type:           "sub",
		PlanSKU:        planSKU,
		StartDate:      start,
		ExpiresAt:      subReq.ExpiresAt,
		CancelledAt:    nil,
		LastModifiedAt: time.Now().Format(time.RFC3339),
		Attributes:     subReq.Metadata,
	}
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = ddbCli.PutItem(ctx, &ddb.PutItemInput{
		TableName:           aws.String(tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(pk) AND attribute_not_exists(sk)"),
	})
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func RenewUserSub(ctx context.Context, ddbCli *ddb.Client, tableName string, subReq *dtos.SubscriptionRequest) error {
	pk, sk := userSubKey(subReq.UserID, subReq.SubscriptionID)
	_, err := ddbCli.UpdateItem(ctx, &ddb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: sk},
		},
		UpdateExpression: aws.String("SET expiresAt = :exp, lastModifiedAt = :now, attributes = :attrs REMOVE canceledAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":exp":   &types.AttributeValueMemberS{Value: subReq.ExpiresAt},
			":now":   &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
			":attrs": parseMetadataAttributes(subReq.Metadata),
		},
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
	})
	if err != nil {
		return fmt.Errorf("failed to renew subscription: %w", err)
	}

	return nil
}

func CancelUserSub(ctx context.Context, ddbCli *ddb.Client, tableName string, subReq *dtos.SubscriptionRequest) error {
	if subReq.CanceledAt == nil || *subReq.CanceledAt == "" {
		return errMissingCanceledAt
	}
	pk, sk := userSubKey(subReq.UserID, subReq.SubscriptionID)
	_, err := ddbCli.UpdateItem(ctx, &ddb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: sk},
		},
		UpdateExpression: aws.String("SET cancelledAt = :canceled, lastModifiedAt = :now, attributes = :attrs"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":canceled": &types.AttributeValueMemberS{Value: *subReq.CanceledAt},
			":now":      &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
			":attrs":    parseMetadataAttributes(subReq.Metadata),
		},
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
	})
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}

func getUserSubs(ctx context.Context, ddbCli *ddb.Client, tableName, userID string) ([]*entities.SubscriptionItem, error) {
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

	subs := make([]*entities.SubscriptionItem, 0, len(out.Items))
	for _, it := range out.Items {
		var s entities.SubscriptionItem
		if err := attributevalue.UnmarshalMap(it, &s); err != nil {
			return nil, fmt.Errorf("failed to unmarshal subscription item: %w", err)
		}
		subs = append(subs, &s)
	}

	return subs, nil
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

func hasActiveOrPending(subs []*entities.SubscriptionItem) (bool, error) {
	for _, s := range subs {
		st, err := getStatus(s.CancelledAt, s.ExpiresAt)
		if err != nil {
			return false, err
		}
		if st == dtos.SubStatusActive || st == dtos.SubStatusPending {
			return true, nil
		}
	}
	return false, nil
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func userSubKey(userID, subID string) (pk, sk string) {
	return "user:" + userID, "sub:" + subID
}

func parseMetadataAttributes(m map[string]any) types.AttributeValue {
	av, err := attributevalue.Marshal(m)
	if err != nil {
		return &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}
	}
	return av
}
