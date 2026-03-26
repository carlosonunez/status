package tokenstore

import (
	"io/fs"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewFileStoreForTest creates a file store with injectable read/write functions
// so unit tests can use in-memory storage without touching the filesystem.
func NewFileStoreForTest(
	readFn func(string) ([]byte, error),
	writeFn func(string, []byte, fs.FileMode) error,
) Store {
	return &fileStore{
		path:      "/test/tokens.json",
		readFile:  readFn,
		writeFile: writeFn,
	}
}

// NewDynamoStoreForTest creates a DynamoDB store from an already-configured
// client and table name. Used by tests that connect to DynamoDB Local.
func NewDynamoStoreForTest(client *dynamodb.Client, tableName string) Store {
	return &dynamoStore{client: client, tableName: tableName}
}
