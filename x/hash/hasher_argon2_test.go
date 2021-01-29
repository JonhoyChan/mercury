package hash

import (
	"mercury/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

type defaultArgon2Config struct{}

func (c defaultArgon2Config) Hasher() *config.Hasher {
	return &config.Hasher{
		Argon2: config.HasherArgon2{
			Memory:      8192,
			Iterations:  2,
			Parallelism: 4,
			SaltLength:  16,
			KeyLength:   16,
		},
	}
}

func TestHasherArgon2(t *testing.T) {
	c := defaultArgon2Config{}
	data := []byte("123456")
	hasher := NewHasherArgon2(c)

	t.Run("hash", func(t *testing.T) {
		hash, err := hasher.Hash(data)
		assert.NoError(t, err)

		t.Logf("argon2: %s", hash)
	})

	t.Run("compare", func(t *testing.T) {
		hash, err := hasher.Hash(data)
		assert.NoError(t, err)

		err = hasher.Compare(data, hash)
		assert.NoError(t, err)
	})
}
