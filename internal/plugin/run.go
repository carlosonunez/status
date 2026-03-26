package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

// invoker calls a plugin binary with a single flag and returns its stdout.
// It is the seam used to inject fakes in tests.
type invoker func(ctx context.Context, path, flag string, stdin io.Reader) ([]byte, error)

// execInvoker is the production invoker that uses os/exec.
func execInvoker(ctx context.Context, path, flag string, stdin io.Reader) ([]byte, error) {
	cmd := exec.CommandContext(ctx, path, flag)
	cmd.Stdin = stdin
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("exec %q %s: %w", path, flag, err)
	}
	return out, nil
}

// fetchMetadata invokes path --metadata using inv and decodes the response.
func fetchMetadata(path string, inv invoker) (metadataResponse, error) {
	out, err := inv(context.Background(), path, "--metadata", bytes.NewReader(nil))
	if err != nil {
		return metadataResponse{}, err
	}
	var meta metadataResponse
	if err := json.Unmarshal(out, &meta); err != nil {
		return metadataResponse{}, fmt.Errorf("decode metadata from %q: %w", path, err)
	}
	return meta, nil
}
