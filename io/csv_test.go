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
		{"header1": "valueA1", "header2": "valueA2"},
		{"header1": "valueB1", "header2": "valueB2"},
	}
	assert.Equal(t, expected, actual)
}
