package token

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"outgoing/app/service/auth/api"
	"outgoing/x/config"
	"outgoing/x/ecode"
	"outgoing/x/types"
	"time"
)

// authenticator is a singleton instance of the authenticator.
type authenticator struct {
	hmacSalt     []byte
	lifetime     time.Duration
	serialNumber int
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
		hmacSalt:     tokenConfig.Key,
		lifetime:     time.Duration(tokenConfig.Expire) * time.Second,
		serialNumber: tokenConfig.SerialNumber,
	}, nil
}

// tokenLayout defines positioning of various bytes in token.
type tokenLayout struct {
	// User ID.
	Uid uint64
	// Token expiration time.
	Expires uint32
	// Refresh Token expiration time.
	RefreshExpires uint32
	// User's authentication level.
	AuthLevel uint16
	// Serial number - to invalidate all tokens if needed.
	SerialNumber uint16
}

func (a *authenticator) generateToken(record *api.Record) (string, error) {
	lifetime, _ := time.ParseDuration(record.Lifetime)
	if lifetime == 0 {
		lifetime = a.lifetime
	} else if lifetime < 0 {
		return "", ecode.ErrTokenExpired
	}
	expires := time.Now().Add(lifetime).UTC().Round(time.Millisecond)

	uid := types.ParseUserUID(record.Uid)

	layout := tokenLayout{
		Uid:            uint64(uid),
		Expires:        uint32(expires.Unix()),
		RefreshExpires: uint32(expires.Add(lifetime).UTC().Round(time.Millisecond).Unix()),
		AuthLevel:      uint16(record.Level),
		SerialNumber:   uint16(a.serialNumber),
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, &layout)
	if err != nil {
		return "", ecode.ErrMalformed
	}
	hasher := hmac.New(sha256.New, a.hmacSalt)
	hasher.Write(buf.Bytes())
	err = binary.Write(buf, binary.LittleEndian, hasher.Sum(nil))
	if err != nil {
		return "", ecode.ErrMalformed
	}

	return hex.EncodeToString(buf.Bytes()), nil
}

// GenerateToken generates a new token.
func (a *authenticator) GenerateToken(record *api.Record) (string, error) {
	return a.generateToken(record)
}

// Authenticate checks validity of provided token.
func (a *authenticator) Authenticate(token string) (*api.Record, string, error) {
	tokenByte, err := hex.DecodeString(token)
	if err != nil {
		return nil, "", ecode.ErrInvalidToken
	}

	var layout tokenLayout
	dataSize := binary.Size(&layout)
	if len(token) < dataSize+sha256.Size {
		// Token is too short
		return nil, "", ecode.ErrInvalidToken
	}

	buf := bytes.NewBuffer(tokenByte)
	err = binary.Read(buf, binary.LittleEndian, &layout)
	if err != nil {
		return nil, "", ecode.ErrInvalidToken
	}

	hBuf := new(bytes.Buffer)
	err = binary.Write(hBuf, binary.LittleEndian, &layout)
	if err != nil {
		return nil, "", ecode.ErrInvalidToken
	}

	// Check signature.
	hasher := hmac.New(sha256.New, a.hmacSalt)
	hasher.Write(hBuf.Bytes())
	if !hmac.Equal(tokenByte[dataSize:dataSize+sha256.Size], hasher.Sum(nil)) {
		return nil, "", ecode.ErrInvalidToken
	}

	// Check authentication level for validity.
	level := api.AuthLevel(layout.AuthLevel)
	if level > api.AuthLevel_AuthLevelRoot {
		return nil, "", ecode.ErrInvalidToken
	}

	// Check serial number.
	if int(layout.SerialNumber) != a.serialNumber {
		return nil, "", ecode.ErrInvalidToken
	}

	uid := types.Uid(layout.Uid).UID()
	record := &api.Record{
		Uid:   uid,
		Level: level,
		State: api.UserState_UserStateNormal,
	}

	// Check token expiration time.
	expires := time.Unix(int64(layout.Expires), 0).UTC()
	now := time.Now().Add(1 * time.Second)
	if expires.Before(now) {
		// Check token refresh expiration time.
		refreshExpires := time.Unix(int64(layout.RefreshExpires), 0).UTC()
		if refreshExpires.Before(now) {
			return nil, "", ecode.ErrTokenExpired
		}

		token, err = a.generateToken(record)
		return nil, token, nil
	}

	record.Lifetime = time.Until(expires).String()
	return record, "", nil
}
