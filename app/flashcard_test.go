package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFromCSV(t *testing.T) {
	lc := LoadConfig{
		Filepath:      "testdata/test.tsv",
		Delimiter:     '\t',
		PromptHeader:  "prompt",
		ContextHeader: "context",
		AnswerHeader:  "answer",
	}
	actual, err := LoadFromCSV(lc)
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := []*Flashcard{
		{Prompt: "P1", Context: "C1", Answers: []string{"A1", "A2"}},
		{Prompt: "P1", Context: "C2", Answers: []string{"A3"}},
		{Prompt: "P2", Answers: []string{"A4"}},
	}
	assert.Equal(t, expected, actual)
}
