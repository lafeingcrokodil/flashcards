package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPercent(t *testing.T) {
	testCases := []struct {
		numerator   int
		denominator int
		expected    int
	}{
		{numerator: 0, denominator: 0, expected: 0},
		{numerator: 1, denominator: 3, expected: 33},
	}

	for _, tc := range testCases {
		actual := Percent(tc.numerator, tc.denominator)
		assert.Equal(t, tc.expected, actual)
	}
}
