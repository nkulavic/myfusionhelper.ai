package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Result contains the outcome of a rate limit check.
type Result struct {
	Allowed bool   `json:"allowed"`
	Limit   int    `json:"limit"`
	Used    int    `json:"used"`
	ResetAt string `json:"reset_at,omitempty"`
}

// Limiter provides rate limiting backed by DynamoDB.
type Limiter struct {
	db              *dynamodb.Client
	accountsTable   string
	rateLimitsTable string
}

// New creates a rate limiter.
func New(db *dynamodb.Client, accountsTable, rateLimitsTable string) *Limiter {
	return &Limiter{
		db:              db,
		accountsTable:   accountsTable,
		rateLimitsTable: rateLimitsTable,
	}
}

// CheckMonthlyLimit checks the account's monthly execution count against plan limit.
// It uses an atomic increment on the account's usage.monthly_executions field.
func (l *Limiter) CheckMonthlyLimit(ctx context.Context, accountID string, maxExecutions int) (*Result, error) {
	if maxExecutions <= 0 {
		// No limit configured
		return &Result{Allowed: true, Limit: 0, Used: 0}, nil
	}

	// Atomic increment of monthly_executions and return new value
	result, err := l.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(l.accountsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
		},
		UpdateExpression: aws.String("SET usage.monthly_executions = if_not_exists(usage.monthly_executions, :zero) + :one"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":zero": &ddbtypes.AttributeValueMemberN{Value: "0"},
			":one":  &ddbtypes.AttributeValueMemberN{Value: "1"},
		},
		ReturnValues: ddbtypes.ReturnValueUpdatedNew,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to increment monthly executions: %w", err)
	}

	// Extract new count
	used := 0
	if v, ok := result.Attributes["usage"]; ok {
		var usage struct {
			MonthlyExecutions int `dynamodbav:"monthly_executions"`
		}
		if m, ok := v.(*ddbtypes.AttributeValueMemberM); ok {
			_ = attributevalue.UnmarshalMap(m.Value, &usage)
			used = usage.MonthlyExecutions
		}
	}

	// Calculate reset time (first of next month)
	now := time.Now().UTC()
	nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	if used > maxExecutions {
		// Rolled back: decrement since we won't actually execute
		_, _ = l.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(l.accountsTable),
			Key: map[string]ddbtypes.AttributeValue{
				"account_id": &ddbtypes.AttributeValueMemberS{Value: accountID},
			},
			UpdateExpression: aws.String("SET usage.monthly_executions = usage.monthly_executions - :one"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":one": &ddbtypes.AttributeValueMemberN{Value: "1"},
			},
		})

		return &Result{
			Allowed: false,
			Limit:   maxExecutions,
			Used:    used - 1, // Show pre-increment value
			ResetAt: nextMonth.Format(time.RFC3339),
		}, nil
	}

	return &Result{
		Allowed: true,
		Limit:   maxExecutions,
		Used:    used,
		ResetAt: nextMonth.Format(time.RFC3339),
	}, nil
}

// CheckBurstLimit checks per-helper per-minute burst rate.
// Uses a DynamoDB item with TTL for auto-cleanup.
func (l *Limiter) CheckBurstLimit(ctx context.Context, helperID string, maxPerMinute int) (*Result, error) {
	if maxPerMinute <= 0 {
		maxPerMinute = 100 // Default burst limit
	}

	now := time.Now().UTC()
	minuteBucket := now.Unix() / 60
	key := fmt.Sprintf("rl:%s:%d", helperID, minuteBucket)
	ttl := now.Add(2 * time.Minute).Unix()

	// Atomic increment with TTL
	result, err := l.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(l.rateLimitsTable),
		Key: map[string]ddbtypes.AttributeValue{
			"key": &ddbtypes.AttributeValueMemberS{Value: key},
		},
		UpdateExpression: aws.String("SET #c = if_not_exists(#c, :zero) + :one, #t = :ttl"),
		ExpressionAttributeNames: map[string]string{
			"#c": "count",
			"#t": "ttl",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":zero": &ddbtypes.AttributeValueMemberN{Value: "0"},
			":one":  &ddbtypes.AttributeValueMemberN{Value: "1"},
			":ttl":  &ddbtypes.AttributeValueMemberN{Value: strconv.FormatInt(ttl, 10)},
		},
		ReturnValues: ddbtypes.ReturnValueUpdatedNew,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check burst limit: %w", err)
	}

	count := 0
	if v, ok := result.Attributes["count"]; ok {
		if n, ok := v.(*ddbtypes.AttributeValueMemberN); ok {
			count, _ = strconv.Atoi(n.Value)
		}
	}

	if count > maxPerMinute {
		return &Result{
			Allowed: false,
			Limit:   maxPerMinute,
			Used:    count,
		}, nil
	}

	return &Result{
		Allowed: true,
		Limit:   maxPerMinute,
		Used:    count,
	}, nil
}
