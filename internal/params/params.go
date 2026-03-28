package params

// Params provides type-safe, read-only access to key/value parameters.
// Implementations are returned by FromMap; callers should not construct
// them directly.
type Params interface {
	// String returns the value for key as a string.
	// Returns ("", false) if the key is absent or the value is not a string.
	String(key string) (string, bool)

	// Bool returns the value for key as a bool.
	// Returns (false, false) if the key is absent or the value is not a bool.
	Bool(key string) (bool, bool)

	// Value returns the raw value for key.
	// Returns (nil, false) if the key is absent.
	// Prefer String or Bool when the expected type is known.
	Value(key string) (any, bool)

	// Keys returns all keys present in the params.
	Keys() []string

	// Has reports whether key exists.
	Has(key string) bool
}

// FromMap wraps m in a Params. If m is nil an empty Params is returned.
func FromMap(m map[string]any) Params {
	if m == nil {
		return &mapParams{data: map[string]any{}}
	}
	// Copy to prevent external mutation.
	cp := make(map[string]any, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return &mapParams{data: cp}
}

type mapParams struct {
	data map[string]any
}

func (p *mapParams) String(key string) (string, bool) {
	v, ok := p.data[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func (p *mapParams) Bool(key string) (bool, bool) {
	v, ok := p.data[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func (p *mapParams) Value(key string) (any, bool) {
	v, ok := p.data[key]
	return v, ok
}

func (p *mapParams) Keys() []string {
	keys := make([]string, 0, len(p.data))
	for k := range p.data {
		keys = append(keys, k)
	}
	return keys
}

func (p *mapParams) Has(key string) bool {
	_, ok := p.data[key]
	return ok
}
