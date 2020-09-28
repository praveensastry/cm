package parser_test

import (
	"testing"

	"github.com/praveensastry/cm/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestGetSpecs(t *testing.T) {
	_, err := parser.GetSpecs()
	assert.NoError(t, err)
}
