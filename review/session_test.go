package review

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPop(t *testing.T) {
	testCases := []struct {
		queue             []int
		numElemsToPop     int
		expectedPopped    []int
		expectedRemaining []int
	}{
		{
			queue:             []int{1, 2, 3, 4, 5},
			numElemsToPop:     3,
			expectedPopped:    []int{1, 2, 3},
			expectedRemaining: []int{4, 5},
		},
		{
			queue:             []int{1, 2, 3, 4, 5},
			numElemsToPop:     6,
			expectedPopped:    []int{1, 2, 3, 4, 5},
			expectedRemaining: nil,
		},
		{
			queue:             []int{},
			numElemsToPop:     1,
			expectedPopped:    []int{},
			expectedRemaining: nil,
		},
		{
			queue:             nil,
			numElemsToPop:     1,
			expectedPopped:    nil,
			expectedRemaining: nil,
		},
	}

	for _, tc := range testCases {
		actualPopped, actualRemaining := pop(tc.queue, tc.numElemsToPop)
		assert.Equal(t, tc.expectedPopped, actualPopped)
		assert.Equal(t, tc.expectedRemaining, actualRemaining)
	}
}
