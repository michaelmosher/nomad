package api

import (
	"reflect"
	"testing"

	"github.com/hashicorp/nomad/api/internal/testutil"
	"github.com/kr/pretty"
)

func TestResources_Canonicalize(t *testing.T) {
	testutil.Parallel(t)
	testCases := []struct {
		name     string
		input    *Resources
		expected *Resources
	}{
		{
			name:     "empty",
			input:    &Resources{},
			expected: DefaultResources(),
		},
		{
			name: "cores",
			input: &Resources{
				Cores:    pointerOf(2),
				MemoryMB: pointerOf(1024),
			},
			expected: &Resources{
				CPU:      pointerOf(0),
				Cores:    pointerOf(2),
				MemoryMB: pointerOf(1024),
			},
		},
		{
			name: "cpu",
			input: &Resources{
				CPU:      pointerOf(500),
				MemoryMB: pointerOf(1024),
			},
			expected: &Resources{
				CPU:      pointerOf(500),
				Cores:    pointerOf(0),
				MemoryMB: pointerOf(1024),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.input.Canonicalize()
			if !reflect.DeepEqual(tc.input, tc.expected) {
				t.Fatalf("Name: %v, Diffs:\n%v", tc.name, pretty.Diff(tc.expected, tc.input))
			}
		})
	}
}
