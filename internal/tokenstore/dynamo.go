package tokenstore

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	internalconfig "github.com/carlosonunez/status/internal/config"
)

const (
	attrService = "service"
	attrKey     = "key"
	attrValue   = "value"
)

type dynamoStore struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoStore creates a DynamoDB-backed Store using cfg.
// The endpoint is resolved from AWS_ENDPOINT_DYNAMODB, then
// cfg.Endpoint, then the AWS SDK default.
func NewDynamoStore(cfg internalconfig.DynamoStoreConfig) (Store, error) {
	endpoint := os.Getenv("AWS_ENDPOINT_DYNAMODB")
	if endpoint == "" {
		endpoint = cfg.Endpoint
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("tokenstore: load aws config: %w", err)
	}

	var opts []func(*dynamodb.Options)
	if endpoint != "" {
		opts = append(opts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	return &dynamoStore{
		client:    dynamodb.NewFromConfig(awsCfg, opts...),
		tableName: cfg.TableName,
	}, nil
}

func (d *dynamoStore) Get(service, key string) (string, error) {
	out, err := d.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			attrService: &types.AttributeValueMemberS{Value: service},
			attrKey:     &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return "", fmt.Errorf("tokenstore: dynamo get: %w", err)
	}
	if out.Item == nil {
		return "", fmt.Errorf("tokenstore: no key %q for service %q", key, service)
	}
	v, ok := out.Item[attrValue].(*types.AttributeValueMemberS)
	if !ok {
		return "", fmt.Errorf("tokenstore: unexpected value type for %q/%q", service, key)
	}
	return v.Value, nil
}

func (d *dynamoStore) Set(service, key, value string) error {
	_, err := d.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item: map[string]types.AttributeValue{
			attrService: &types.AttributeValueMemberS{Value: service},
			attrKey:     &types.AttributeValueMemberS{Value: key},
			attrValue:   &types.AttributeValueMemberS{Value: value},
		},
	})
	if err != nil {
		return fmt.Errorf("tokenstore: dynamo put: %w", err)
	}
	return nil
}

func (d *dynamoStore) Delete(service, key string) error {
	_, err := d.client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			attrService: &types.AttributeValueMemberS{Value: service},
			attrKey:     &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return fmt.Errorf("tokenstore: dynamo delete: %w", err)
	}
	return nil
}
