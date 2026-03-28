package tokenstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

type fileStore struct {
	mu        sync.Mutex
	path      string
	readFile  func(string) ([]byte, error)
	writeFile func(string, []byte, fs.FileMode) error
}

// NewFileStore creates a file-backed Store that persists tokens at path.
func NewFileStore(path string) Store {
	return &fileStore{
		path:      path,
		readFile:  os.ReadFile,
		writeFile: os.WriteFile,
	}
}

func (f *fileStore) load() (map[string]map[string]string, error) {
	data, err := f.readFile(f.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]map[string]string), nil
		}
		return nil, fmt.Errorf("tokenstore: read %s: %w", f.path, err)
	}
	var m map[string]map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("tokenstore: parse %s: %w", f.path, err)
	}
	if m == nil {
		m = make(map[string]map[string]string)
	}
	return m, nil
}

func (f *fileStore) save(m map[string]map[string]string) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("tokenstore: marshal: %w", err)
	}
	if err := f.writeFile(f.path, data, 0o600); err != nil {
		return fmt.Errorf("tokenstore: write %s: %w", f.path, err)
	}
	return nil
}

func (f *fileStore) Get(service, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, err := f.load()
	if err != nil {
		return "", err
	}
	svc, ok := m[service]
	if !ok {
		return "", fmt.Errorf("tokenstore: no tokens for service %q", service)
	}
	val, ok := svc[key]
	if !ok {
		return "", fmt.Errorf("tokenstore: no key %q for service %q", key, service)
	}
	return val, nil
}

func (f *fileStore) Set(service, key, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, err := f.load()
	if err != nil {
		return err
	}
	if m[service] == nil {
		m[service] = make(map[string]string)
	}
	m[service][key] = value
	return f.save(m)
}

func (f *fileStore) Delete(service, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, err := f.load()
	if err != nil {
		return err
	}
	if svc, ok := m[service]; ok {
		delete(svc, key)
		if len(svc) == 0 {
			delete(m, service)
		}
	}
	return f.save(m)
}
