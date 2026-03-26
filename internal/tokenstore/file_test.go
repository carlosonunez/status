package tokenstore_test

import (
	"io/fs"
	"os"
	"testing"

	"github.com/carlosonunez/status/internal/tokenstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memPersist returns a matched read/write pair backed by a single byte slice.
// The returned functions share state so writes are visible to subsequent reads.
func memPersist() (func(string) ([]byte, error), func(string, []byte, fs.FileMode) error) {
	var data []byte
	read := func(_ string) ([]byte, error) {
		if data == nil {
			return nil, os.ErrNotExist
		}
		out := make([]byte, len(data))
		copy(out, data)
		return out, nil
	}
	write := func(_ string, b []byte, _ fs.FileMode) error {
		data = make([]byte, len(b))
		copy(data, b)
		return nil
	}
	return read, write
}

type fileStoreTest struct {
	TestName string
	Run      func(t *testing.T, s tokenstore.Store)
}

func (tc fileStoreTest) RunTest(t *testing.T) {
	t.Helper()
	read, write := memPersist()
	s := tokenstore.NewFileStoreForTest(read, write)
	tc.Run(t, s)
}

func TestFileStore(t *testing.T) {
	tests := []fileStoreTest{
		{
			TestName: "get_missing_service_returns_error",
			Run: func(t *testing.T, s tokenstore.Store) {
				_, err := s.Get("slack", "access_token")
				require.Error(t, err)
			},
		},
		{
			TestName: "get_missing_key_returns_error",
			Run: func(t *testing.T, s tokenstore.Store) {
				require.NoError(t, s.Set("slack", "access_token", "xoxp-123"))
				_, err := s.Get("slack", "refresh_token")
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
			TestName: "delete_removes_key",
			Run: func(t *testing.T, s tokenstore.Store) {
				require.NoError(t, s.Set("slack", "access_token", "xoxp-123"))
				require.NoError(t, s.Delete("slack", "access_token"))
				_, err := s.Get("slack", "access_token")
				require.Error(t, err)
			},
		},
		{
			TestName: "delete_nonexistent_key_is_a_noop",
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
		{
			TestName: "delete_last_key_removes_service_entry",
			Run: func(t *testing.T, s tokenstore.Store) {
				require.NoError(t, s.Set("slack", "access_token", "tok"))
				require.NoError(t, s.Delete("slack", "access_token"))
				// Service entry itself should be gone; a subsequent Set should still work.
				require.NoError(t, s.Set("slack", "access_token", "new-tok"))
				val, _ := s.Get("slack", "access_token")
				assert.Equal(t, "new-tok", val)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}
