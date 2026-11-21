package interfaces

import (
	"testing"

	"slices"

	"github.com/osteele/liquid"
)

func TestRenderString(t *testing.T) {
	engine := liquid.NewEngine()
	cases := []struct {
		key      any
		expected string
		bindings map[string]any
	}{
		{
			key:      nil,
			expected: "",
			bindings: make(map[string]any),
		},
		{
			key:      "",
			expected: "",
			bindings: make(map[string]any),
		},
		{
			key:      "value",
			expected: "value",
			bindings: make(map[string]any),
		},
		{
			key:      "test.{{ liquid }}",
			expected: "test.hello",
			bindings: map[string]any{
				"liquid": "hello",
			},
		},
	}

	for _, testcase := range cases {
		result := RenderString(testcase.key, engine, testcase.bindings)
		if result != testcase.expected {
			t.Errorf("result does not match test: %case", testcase.key)
		}
	}
}

func TestRenderStringArray(t *testing.T) {
	engine := liquid.NewEngine()
	cases := []struct {
		key      any
		expected []string
		bindings map[string]any
	}{
		{
			key:      nil,
			expected: nil,
			bindings: make(map[string]any),
		},
		{
			key:      "",
			expected: []string{""},
			bindings: make(map[string]any),
		},
		{
			key:      []string{"value1", "value2", "value3"},
			expected: []string{"value1", "value2", "value3"},
			bindings: make(map[string]any),
		},
		{
			key:      []string{"value1", "{{liquid}}", "{{liquid}} {{val}}"},
			expected: []string{"value1", "hello", "hello world"},
			bindings: map[string]any{
				"liquid": "hello",
				"val":    "world",
			},
		},
	}

	for _, testcase := range cases {
		result := RenderStringArray(testcase.key, engine, testcase.bindings)
		if testcase.expected == nil {
			if result != nil {
				t.Errorf("result does not match test: %case", testcase.key)
			}
		}

		if !slices.Equal(result, testcase.expected) {
			t.Errorf("result does not match test: %case", testcase.key)
		}
	}
}
