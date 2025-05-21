package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadAllTSV(t *testing.T) {
	actual, err := ReadAllTSV("testdata/test.tsv")
	if err != nil {
		t.Fatal(err.Error())
	}
	expected := []map[string]string{
		{"header1": "valueA1", "header2": "valueA2"},
		{"header1": "valueB1", "header2": "valueB2"},
	}
	assert.Equal(t, expected, actual)
}
