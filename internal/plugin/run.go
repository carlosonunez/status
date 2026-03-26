package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

// runBinary executes the binary at path with the given flag, piping stdin into
// it and returning its stdout. Stderr is discarded.
func runBinary(ctx context.Context, path, flag string, stdin io.Reader) ([]byte, error) {
	cmd := exec.CommandContext(ctx, path, flag)
	cmd.Stdin = stdin
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("exec %q %s: %w", path, flag, err)
	}
	return out, nil
}

// fetchMetadata invokes path --metadata and decodes the response.
func fetchMetadata(path string) (metadataResponse, error) {
	ctx := context.Background()
	out, err := runBinary(ctx, path, "--metadata", bytes.NewReader(nil))
	if err != nil {
		return metadataResponse{}, err
	}
	var meta metadataResponse
	if err := json.Unmarshal(out, &meta); err != nil {
		return metadataResponse{}, fmt.Errorf("decode metadata from %q: %w", path, err)
	}
	return meta, nil
}
