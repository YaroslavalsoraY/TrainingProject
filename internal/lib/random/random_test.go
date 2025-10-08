package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandom(t *testing.T) {
	tests := []struct{length int}{
		{length: 1},
		{length: 2},
		{length: 4},
		{length: 8},
		{length: 16},
	}

	for _, tt := range tests {
		result := NewRandomString(tt.length)
		assert.Equal(t, len(result), tt.length)
	}
}