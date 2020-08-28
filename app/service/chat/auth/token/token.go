package token

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"outgoing/x/config"
	"outgoing/x/ecode"
	"time"
)

type Authenticator interface {
	GenerateToken(data interface{}) (token string, lifetime string, err error)

	Authenticate(token string, out interface{}) (lifetime string, err error)
}

// authenticator is a singleton instance of the authenticator.
type authenticator struct {
	hmacSalt []byte
	lifetime time.Duration
}

type AuthenticatorTokenProvider interface {
	AuthenticatorToken() *config.AuthenticatorTokenConfig
}

func NewAuthenticator(p AuthenticatorTokenProvider) (*authenticator, error) {
	tokenConfig := p.AuthenticatorToken()

	if len(tokenConfig.Key) < sha256.Size {
		return nil, ecode.NewError("the key is missing or too short")
	}
	if tokenConfig.Expire <= 0 {
		return nil, ecode.NewError("invalid expiration value")
	}

	return &authenticator{
		hmacSalt: tokenConfig.Key,
		lifetime: time.Duration(tokenConfig.Expire) * time.Second,
	}, nil
}

// tokenLayout defines positioning of various bytes in token.
type tokenLayout struct {
	// Token expiration time.
	Expires uint32
	// Token data
	Data []byte
}

func (a *authenticator) generateToken(data interface{}) (string, string, error) {
	lifetime := a.lifetime
	if lifetime < 0 {
		return "", "", ecode.ErrTokenExpired
	}
	expires := time.Now().Add(lifetime).UTC().Round(time.Second)

	layout := tokenLayout{
		Expires: uint32(expires.Unix()),
	}

	switch v := data.(type) {
	case string:
		layout.Data = []byte(v)
	default:
		return "", "", fmt.Errorf("unsupported type: %T", v)
	}

	layoutData, err := jsoniter.Marshal(&layout)
	if err != nil {
		return "", "", err
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, layoutData)
	if err != nil {
		return "", "", ecode.ErrMalformed
	}
	hasher := hmac.New(sha256.New, a.hmacSalt)
	hasher.Write(buf.Bytes())
	err = binary.Write(buf, binary.LittleEndian, hasher.Sum(nil))
	if err != nil {
		return "", "", ecode.ErrMalformed
	}

	return hex.EncodeToString(buf.Bytes()), time.Until(expires).String(), nil
}

// GenerateToken generates a new token.
func (a *authenticator) GenerateToken(data interface{}) (string, string, error) {
	return a.generateToken(data)
}

// Authenticate checks validity of provided token.
func (a *authenticator) Authenticate(token string, out interface{}) (string, error) {
	tokenByte, err := hex.DecodeString(token)
	if err != nil {
		return "", err
	}

	if len(tokenByte) <= sha256.Size {
		return "", ecode.ErrInvalidToken
	}

	dataSize := len(tokenByte) - sha256.Size
	layoutData := tokenByte[:dataSize]

	hBuf := new(bytes.Buffer)
	err = binary.Write(hBuf, binary.LittleEndian, &layoutData)
	if err != nil {
		return "", err
	}

	// Check signature.
	hasher := hmac.New(sha256.New, a.hmacSalt)
	hasher.Write(hBuf.Bytes())
	if !hmac.Equal(tokenByte[dataSize:dataSize+sha256.Size], hasher.Sum(nil)) {
		return "", ecode.ErrInvalidToken
	}

	var layout tokenLayout
	err = jsoniter.Unmarshal(layoutData, &layout)
	if err != nil {
		return "", err
	}

	switch v := out.(type) {
	case *string:
		*v = string(layout.Data)
	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}

	// Check token expiration time.
	expires := time.Unix(int64(layout.Expires), 0).UTC()
	now := time.Now().Add(1 * time.Second)
	if expires.Before(now) {
		return "", ecode.ErrTokenExpired
	}

	return time.Until(expires).String(), nil
}
