package jwt

import (
	"testing"
	"outgoing/app/service/main/auth/api"
	"outgoing/x/config"

	"github.com/stretchr/testify/assert"
)

type defaultTokenConfig struct{}

func (c defaultTokenConfig) AuthenticatorJWT() *config.AuthenticatorJWTConfig {
	return &config.AuthenticatorJWTConfig{
		Expire:       1209600,
		SerialNumber: 1,
		Key:          []byte("AG5s4d68asg7SF5sdf454ghj"),
	}
}

func TestAuthenticator_GenerateToken(t *testing.T) {
	c := defaultTokenConfig{}

	handler, err := NewAuthenticator(c, nil)
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
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJEYXRhIjoiMk8zM3hRQ0xLQWJOL2daZnpYTVpYd01BQVFBPSIsImV4cCI6MTU5NDI5Mzk2NX0.I2XtsPrK0vI8VvdyB75q6D2RMeLP9SiOl84wyN5px1E"

	handler, err := NewAuthenticator(c, nil)
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
