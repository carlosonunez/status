package params_test

import (
	"testing"

	"github.com/carlosonunez/status/internal/params"
	"github.com/stretchr/testify/assert"
)

type stringTest struct {
	TestName string
	Key      string
	Input    map[string]any
	WantVal  string
	WantOK   bool
}

func (tc *stringTest) Run(t *testing.T) {
	t.Helper()
	p := params.FromMap(tc.Input)
	got, ok := p.String(tc.Key)
	assert.Equal(t, tc.WantOK, ok, "[%s] ok mismatch", tc.TestName)
	assert.Equal(t, tc.WantVal, got, "[%s] value mismatch", tc.TestName)
}

func TestParams_String(t *testing.T) {
	tests := []stringTest{
		{TestName: "present_string", Key: "msg", Input: map[string]any{"msg": "hello"}, WantVal: "hello", WantOK: true},
		{TestName: "absent_key", Key: "missing", Input: map[string]any{}, WantVal: "", WantOK: false},
		{TestName: "wrong_type", Key: "n", Input: map[string]any{"n": 42}, WantVal: "", WantOK: false},
		{TestName: "nil_map", Key: "x", Input: nil, WantVal: "", WantOK: false},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, tc.Run)
	}
}

type boolTest struct {
	TestName string
	Key      string
	Input    map[string]any
	WantVal  bool
	WantOK   bool
}

func (tc *boolTest) Run(t *testing.T) {
	t.Helper()
	p := params.FromMap(tc.Input)
	got, ok := p.Bool(tc.Key)
	assert.Equal(t, tc.WantOK, ok, "[%s] ok mismatch", tc.TestName)
	assert.Equal(t, tc.WantVal, got, "[%s] value mismatch", tc.TestName)
}

func TestParams_Bool(t *testing.T) {
	tests := []boolTest{
		{TestName: "present_true", Key: "ooo", Input: map[string]any{"ooo": true}, WantVal: true, WantOK: true},
		{TestName: "present_false", Key: "ooo", Input: map[string]any{"ooo": false}, WantVal: false, WantOK: true},
		{TestName: "absent_key", Key: "missing", Input: map[string]any{}, WantVal: false, WantOK: false},
		{TestName: "wrong_type", Key: "s", Input: map[string]any{"s": "yes"}, WantVal: false, WantOK: false},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, tc.Run)
	}
}

func TestParams_Value(t *testing.T) {
	p := params.FromMap(map[string]any{"n": 42, "s": "hello"})
	v, ok := p.Value("n")
	assert.True(t, ok)
	assert.Equal(t, 42, v)

	_, ok = p.Value("missing")
	assert.False(t, ok)
}

func TestParams_Has(t *testing.T) {
	p := params.FromMap(map[string]any{"a": 1})
	assert.True(t, p.Has("a"))
	assert.False(t, p.Has("b"))
}

func TestParams_Keys(t *testing.T) {
	p := params.FromMap(map[string]any{"x": 1, "y": 2, "z": 3})
	keys := p.Keys()
	assert.Len(t, keys, 3)
	assert.ElementsMatch(t, []string{"x", "y", "z"}, keys)
}

func TestParams_FromMap_NilProducesEmptyParams(t *testing.T) {
	p := params.FromMap(nil)
	assert.Empty(t, p.Keys())
	assert.False(t, p.Has("anything"))
}
