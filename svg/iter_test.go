package svg

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
)

var eq = qt.CmpEquals(
	cmp.Comparer(func(i1, i2 XyIndex) bool {
		return i1.X == i2.X && i1.Y == i2.Y
	}),
)

func TestIterators(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		iterator chan XyIndex
		expected []XyIndex
	}{
		// UpDown
		// 1 3
		// 2 4
		{
			iterator: UpDownMinor(2, 2),
			expected: []XyIndex{
				{0, 0},
				{0, 1},
				{1, 0},
				{1, 1},
			},
		},

		// LeftRightMinor
		// 1 2
		// 3 4
		{
			iterator: LeftRightMinor(2, 2),
			expected: []XyIndex{
				{0, 0},
				{1, 0},
				{0, 1},
				{1, 1},
			},
		},

		// DiagUp
		// 1 3
		// 2 5
		// 4 6
		{
			iterator: DiagUp(2, 3),
			expected: []XyIndex{
				{0, 0}, // x + y == 0
				{0, 1}, // x + y == 1
				{1, 0}, // x + y == 1
				{0, 2}, // x + y == 2
				{1, 1}, // x + y == 2
				{1, 2}, // x + y == 3
			},
		},

		// DiagDown
		// 2 4 6
		// 1 3 5
		{
			iterator: DiagDown(3, 2),
			expected: []XyIndex{
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
		result := make([]XyIndex, 0, len(tt.expected))

		for i := range tt.iterator {
			result = append(result, i)
		}

		c.Assert(result, eq, tt.expected)
	}
}
