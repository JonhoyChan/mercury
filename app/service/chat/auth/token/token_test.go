package token

import (
	"outgoing/x/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

type defaultTokenConfig struct{}

func (c defaultTokenConfig) AuthenticatorToken() *config.AuthenticatorTokenConfig {
	return &config.AuthenticatorTokenConfig{
		Expire:       1209600,
		SerialNumber: 1,
		Key:          []byte("wfaY2RgF2S1OQI/ZlK+LSrp1KB2jwAdGAIHQ7JZn+Kc="),
	}
}

func TestToken(t *testing.T) {
	c := defaultTokenConfig{}

	handler, err := NewAuthenticator(c)
	assert.NoError(t, err)

	clientID := "666666"
	t.Run("generate token", func(t *testing.T) {
		token, lifetime, err := handler.GenerateToken(clientID)
		assert.NoError(t, err)

		t.Logf("token: %v, lifetime: %v \n", token, lifetime)
	})

	t.Run("authenticate", func(t *testing.T) {
		token := "7b2245787069726573223a313539393631393530392c2244617461223a224e6a59324e6a5932227dbe1c5a54525d71306f4cafb854342910af1166f0d05489ca0f35e4edd0a4c626"

		var out string
		lifetime, err := handler.Authenticate(token, &out)
		assert.NoError(t, err)

		assert.Equal(t, clientID, out)
		t.Logf("lifetime: %v \n", lifetime)
	})
}
