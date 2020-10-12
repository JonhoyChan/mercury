package jwt

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"mercury/x/ecode"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Authenticator interface {
	GenerateToken(issuer string, key []byte, expire time.Duration, data interface{}) (string, string, error)

	Authenticate(token, issuer string, key []byte, out interface{}) (string, error)
}

type authenticator struct{}

func NewAuthenticator() (*authenticator, error) {
	return &authenticator{}, nil
}

type CustomClaims struct {
	Data []byte
	jwt.StandardClaims
}

type tokenLayout struct {
	// Token expiration time.
	Expires int64
	// Token data
	Data []byte
}

func (a *authenticator) generateToken(issuer string, key []byte, expire time.Duration, data interface{}) (string, string, error) {
	var layout tokenLayout

	switch v := data.(type) {
	case string:
		layout.Data = []byte(v)
	default:
		return "", "", fmt.Errorf("unsupported type: %T", v)
	}

	standardClaims := jwt.StandardClaims{
		Issuer: issuer,
	}

	var lifetime string
	if expire > 0 {
		expires := time.Now().Add(expire).UTC().Round(time.Millisecond)
		layout.Expires = expires.Unix()
		standardClaims.ExpiresAt = expires.Unix()
		lifetime = time.Until(expires).String()
	}

	layoutData, err := jsoniter.Marshal(&layout)
	if err != nil {
		return "", "", err
	}

	claims := CustomClaims{
		Data:           layoutData,
		StandardClaims: standardClaims,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", "", err
	}

	return signedToken, lifetime, nil
}

// GenerateToken generates a new token.
func (a *authenticator) GenerateToken(issuer string, key []byte, expire time.Duration, data interface{}) (string, string, error) {
	return a.generateToken(issuer, key, expire, data)
}

func (a *authenticator) Authenticate(token, issuer string, key []byte, out interface{}) (string, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return key, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors == jwt.ValidationErrorExpired {
				return "", ecode.ErrTokenExpired
			} else {
				return "", ecode.ErrInvalidToken
			}
		} else {
			return "", ecode.ErrInvalidToken
		}
	}

	if jwtToken != nil {
		if claims, ok := jwtToken.Claims.(*CustomClaims); ok {
			// Check token issuer
			if claims.Issuer != issuer {
				return "", ecode.ErrInvalidToken
			}

			var layout tokenLayout
			err = jsoniter.Unmarshal(claims.Data, &layout)
			if err != nil {
				return "", err
			}

			switch v := out.(type) {
			case *string:
				*v = string(layout.Data)
			default:
				return "", fmt.Errorf("unsupported type: %T", v)
			}

			var lifetime string
			if claims.ExpiresAt > 0 {
				expires := time.Unix(claims.ExpiresAt, 0).UTC()
				lifetime = time.Until(expires).String()
			}

			return lifetime, nil
		}
		return "", ecode.ErrInvalidToken
	}

	return "", ecode.ErrInvalidToken
}
