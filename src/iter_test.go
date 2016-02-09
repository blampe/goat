package goat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterators(t *testing.T) {

	tests := []struct {
		iterator chan Index
		expected []Index
	}{
		// UpDown
		// 1 3
		// 2 4
		{
			iterator: upDown(2, 2),
			expected: []Index{
				{0, 0},
				{0, 1},
				{1, 0},
				{1, 1},
			},
		},

		// LeftRight
		// 1 2
		// 3 4
		{
			iterator: leftRight(2, 2),
			expected: []Index{
				{0, 0},
				{1, 0},
				{0, 1},
				{1, 1},
			},
		},

		// DiagUp
		// 1 2
		// 2 3
		// 3 4
		{
			iterator: diagUp(2, 3),
			expected: []Index{
				{0, 0}, // x + y == 0
				{0, 1}, // x + y == 1
				{1, 0}, // x + y == 1
				{0, 2}, // x + y == 2
				{1, 1}, // x + y == 2
				{1, 2}, // x + y == 3
			},
		},

		// DiagDown
		// 3 2 1
		// 4 3 2
		{
			iterator: diagDown(3, 2),
			expected: []Index{
				{0, 1}, // x - y == -1
				{0, 0}, // x - y == 0
				{1, 1}, // x - y == 0
				{1, 0}, // x - y == 1
				{2, 1}, // x - y == 1
				{2, 0}, // x - y == 2
			},
		},
	}

	for _, tt := range tests {
		result := make([]Index, 0, len(tt.expected))

		for i := range tt.iterator {
			result = append(result, i)
		}

		assert.Equal(t, tt.expected, result)
	}
}
