package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	authenticator, _ := NewAuthenticator()

	var (
		issuer = "mercury"
		key    = []byte("AG5s4d68asg7SF5sdf454ghj")
		expire = 60 * time.Second
	)

	var token string
	t.Run("generate token", func(t *testing.T) {
		var err error
		token, _, err = authenticator.GenerateToken(issuer, key, expire, "uidzm74nmfx1O4")
		assert.NoError(t, err)

		t.Logf("generate token: %v", token)
	})

	t.Run("authenticate", func(t *testing.T) {
		var uid string
		lifetime, err := authenticator.Authenticate(token, issuer, key, &uid)
		assert.NoError(t, err)
		assert.Equal(t, uid, "uidzm74nmfx1O4")

		t.Logf("lifetime: %v \n", lifetime)
	})
}
