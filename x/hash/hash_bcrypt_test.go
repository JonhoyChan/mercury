package hash

import (
	"github.com/stretchr/testify/assert"
	"mercury/x/config"
	"testing"
)

type defaultBCryptConfig struct{}

func (c defaultBCryptConfig) HasherBCrypt() *config.HasherBCryptConfig {
	return &config.HasherBCryptConfig{
		Cost: 10,
	}
}

func TestHasherBCrypt(t *testing.T) {
	c := defaultBCryptConfig{}
	data := []byte("123456")
	hasher := NewHasherBCrypt(c)

	t.Run("hash", func(t *testing.T) {
		hash, err := hasher.Hash(data)
		assert.NoError(t, err)

		t.Logf("bcrypt: %s", hash)
	})

	t.Run("compare", func(t *testing.T) {
		hash, err := hasher.Hash(data)
		assert.NoError(t, err)

		err = hasher.Compare(hash, data)
		assert.NoError(t, err)
	})
}
