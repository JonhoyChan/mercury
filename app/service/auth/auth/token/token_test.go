package token

import (
	"outgoing/app/service/auth/api"
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

func TestAuthenticator_GenerateToken(t *testing.T) {
	c := defaultTokenConfig{}

	handler, err := NewAuthenticator(c)
	assert.NoError(t, err)
	record := &api.Record{
		Uid:   "uid2O33xQCLKAY",
		Level: api.AuthLevel_AuthLevelRoot,
		State: api.UserState_UserStateNormal,
	}
	token, err := handler.GenerateToken(record)
	assert.NoError(t, err)

	t.Logf("generate token: %v", token)
}

func TestHasherArgon2_Authenticate(t *testing.T) {
	c := defaultTokenConfig{}
	token := "d8edf7c5008b2806946ee25e030001000000d8ddf43aabaa6fa473c85e8affafd87523c85f6c8c0f5f624fd16b81dd139723"

	handler, err := NewAuthenticator(c)
	assert.NoError(t, err)

	var record *api.Record
	record, token, err = handler.Authenticate(token)
	assert.NoError(t, err)

	if record != nil {
		t.Logf("lifetime: %v", record.Lifetime)

		assert.Equal(t, "uid2O33xQCLKAY", record.Uid)
		assert.Equal(t, api.AuthLevel_AuthLevelRoot, record.Level)
		assert.Equal(t, api.UserState_UserStateNormal, record.State)
	} else if token != "" {
		t.Logf("refresh token: %v", token)
	}
}
