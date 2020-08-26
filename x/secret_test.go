package x

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateSecret(t *testing.T) {
	secret, err := GenerateSecret(26)
	assert.NoError(t, err)

	t.Logf("%s", string(secret))
}
