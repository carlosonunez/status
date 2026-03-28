package tokenstore_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/carlosonunez/status/internal/tokenstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dynamoEndpoint() string {
	if e := os.Getenv("AWS_ENDPOINT_DYNAMODB"); e != "" {
		return e
	}
	return "http://localhost:8000"
}

// newTestDynamoStore creates a DynamoDB client pointed at DynamoDB Local,
// creates a uniquely-named table, registers cleanup, and returns a Store.
// The test is skipped if DynamoDB Local is not reachable.
func newTestDynamoStore(t *testing.T) tokenstore.Store {
	t.Helper()

	endpoint := dynamoEndpoint()
	host := strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")

	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		t.Skipf("DynamoDB Local not reachable at %s: %v", endpoint, err)
	}
	conn.Close()

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("local", "local", ""),
		),
	)
	require.NoError(t, err)

	client := dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	tableName := fmt.Sprintf("tokenstore-test-%d", time.Now().UnixNano())
	_, err = client.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("service"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("key"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("service"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("key"), KeyType: types.KeyTypeRange},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	return tokenstore.NewDynamoStoreForTest(client, tableName)
}

type dynamoStoreTest struct {
	TestName string
	Run      func(t *testing.T, s tokenstore.Store)
}

func (tc dynamoStoreTest) RunTest(t *testing.T) {
	t.Helper()
	s := newTestDynamoStore(t)
	tc.Run(t, s)
}

func TestDynamoStore(t *testing.T) {
	tests := []dynamoStoreTest{
		{
			TestName: "get_missing_item_returns_error",
			Run: func(t *testing.T, s tokenstore.Store) {
				_, err := s.Get("slack", "access_token")
				require.Error(t, err)
			},
		},
		{
			TestName: "set_then_get_returns_value",
			Run: func(t *testing.T, s tokenstore.Store) {
				require.NoError(t, s.Set("slack", "access_token", "xoxp-123"))
				val, err := s.Get("slack", "access_token")
				require.NoError(t, err)
				assert.Equal(t, "xoxp-123", val)
			},
		},
		{
			TestName: "multiple_services_are_isolated",
			Run: func(t *testing.T, s tokenstore.Store) {
				require.NoError(t, s.Set("slack", "access_token", "slack-tok"))
				require.NoError(t, s.Set("google", "access_token", "google-tok"))
				v1, err := s.Get("slack", "access_token")
				require.NoError(t, err)
				v2, err := s.Get("google", "access_token")
				require.NoError(t, err)
				assert.Equal(t, "slack-tok", v1)
				assert.Equal(t, "google-tok", v2)
			},
		},
		{
			TestName: "delete_removes_item",
			Run: func(t *testing.T, s tokenstore.Store) {
				require.NoError(t, s.Set("slack", "access_token", "xoxp-123"))
				require.NoError(t, s.Delete("slack", "access_token"))
				_, err := s.Get("slack", "access_token")
				require.Error(t, err)
			},
		},
		{
			TestName: "delete_nonexistent_item_is_a_noop",
			Run: func(t *testing.T, s tokenstore.Store) {
				require.NoError(t, s.Delete("slack", "nonexistent"))
			},
		},
		{
			TestName: "set_overwrites_existing_value",
			Run: func(t *testing.T, s tokenstore.Store) {
				require.NoError(t, s.Set("slack", "access_token", "old"))
				require.NoError(t, s.Set("slack", "access_token", "new"))
				val, err := s.Get("slack", "access_token")
				require.NoError(t, err)
				assert.Equal(t, "new", val)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}
