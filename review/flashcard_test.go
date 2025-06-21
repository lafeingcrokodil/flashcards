package review

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlashcard_Check(t *testing.T) {
	testCases := []struct {
		prompt    string
		answer    string
		submitted string
		expected  bool
	}{
		{
			prompt:    "P1",
			answer:    "A1",
			submitted: "A1",
			expected:  true,
		},
		{
			prompt:    "P1",
			answer:    "A1",
			submitted: "A2",
			expected:  false,
		},
	}

	for _, tc := range testCases {
		f := &Flashcard{
			Prompt: tc.prompt,
			Answer: tc.answer,
		}
		actual := f.Check(tc.submitted)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestFlashcard_QualifiedPrompt(t *testing.T) {
	testCases := []struct {
		prompt                  string
		context                 string
		expectedQualifiedPrompt string
	}{
		{
			prompt:                  "P1",
			context:                 "C1",
			expectedQualifiedPrompt: "P1 (C1)",
		},
		{
			prompt:                  "P1",
			expectedQualifiedPrompt: "P1",
		},
	}

	for _, tc := range testCases {
		f := &Flashcard{
			Prompt:  tc.prompt,
			Context: tc.context,
		}
		actual := f.QualifiedPrompt()
		assert.Equal(t, tc.expectedQualifiedPrompt, actual)
	}
}
