package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadAllCSV(t *testing.T) {
	actual, err := ReadAllCSV("testdata/test.tsv", '\t')
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := []map[string]string{
		{"A": "A1", "B": "B1"},
		{"A": "A2", "B": "B2"},
	}
	assert.Equal(t, expected, actual)
}
