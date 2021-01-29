package config

import (
	"time"
)

type Authenticator struct {
	Token AuthenticatorToken `json:"token"`
}

type AuthenticatorToken struct {
	Expire       time.Duration `json:"expire"`
	SerialNumber int           `json:"serial_number"`
	Key          []byte        `json:"key"`
}

func DefaultAuthenticator() *Authenticator {
	return &Authenticator{
		Token: AuthenticatorToken{
			// Lifetime of a security token in seconds. 1209600 = 2 weeks.
			Expire: 1209600 * time.Second,
			// Serial number of the token. Can be used to invalidate all issued tokens at once.
			SerialNumber: 1,
			// Secret key (HMAC salt) for signing the tokens.
			Key: []byte("wfaY2RgF2S1OQI/ZlK+LSrp1KB2jwAdGAIHQ7JZn+Kc="),
		},
	}
}
