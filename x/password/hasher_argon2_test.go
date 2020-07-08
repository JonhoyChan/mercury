package password

import (
	"testing"
	"outgoing/x/config"

	"github.com/stretchr/testify/assert"
)

type defaultArgon2Config struct{}

func (c defaultArgon2Config) HasherArgon2() *config.HasherArgon2Config {
	return &config.HasherArgon2Config{
		Memory:      8192,
		Iterations:  2,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   16,
	}
}

func TestHasherArgon2_Generate(t *testing.T) {
	c := defaultArgon2Config{}

	hasher := NewHasherArgon2(c)
	argon2Password, err := hasher.Generate("123456")
	assert.NoError(t, err)

	t.Logf("argon2 password: %s", argon2Password)
}

func TestHasherArgon2_Compare(t *testing.T) {
	c := defaultArgon2Config{}
	password := "123456"

	hasher := NewHasherArgon2(c)
	argon2Password, err := hasher.Generate(password)
	assert.NoError(t, err)

	err = hasher.Compare(password, argon2Password)
	assert.NoError(t, err)
}
