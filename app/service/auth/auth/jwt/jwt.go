package jwt

import (
	"bytes"
	"encoding/binary"
	"outgoing/app/service/auth/api"
	"outgoing/app/service/auth/persistence"
	"outgoing/x/config"
	"outgoing/x/ecode"
	"outgoing/x/types"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type authenticator struct {
	signingKey   []byte
	lifetime     time.Duration
	serialNumber int
	issuer       string
	cacher       persistence.Cacher
}

type AuthenticatorTokenProvider interface {
	AuthenticatorJWT() *config.AuthenticatorJWTConfig
}

func NewAuthenticator(p AuthenticatorTokenProvider, c persistence.Cacher) (*authenticator, error) {
	jwtConfig := p.AuthenticatorJWT()

	if len(jwtConfig.Key) <= 0 {
		return nil, ecode.NewError("the key is missing or too short")
	}
	if jwtConfig.Expire <= 0 {
		return nil, ecode.NewError("invalid expiration value")
	}

	return &authenticator{
		signingKey:   jwtConfig.Key,
		lifetime:     time.Duration(jwtConfig.Expire) * time.Second,
		serialNumber: jwtConfig.SerialNumber,
		issuer:       jwtConfig.Issuer,
		cacher:       c,
	}, nil
}

type CustomClaims struct {
	Data []byte
	jwt.StandardClaims
}

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

	claims := CustomClaims{
		Data: buf.Bytes(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires.Unix(), // 过期时间
			Issuer:    a.issuer,       // 签名的发行者
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(a.signingKey)
	if err != nil {
		return "", ecode.ErrMalformed
	}

	if a.cacher != nil {
		if err := a.cacher.SetAuthToken(record.Uid, signedToken, lifetime); err != nil {
			return "", ecode.ErrInternalServer
		}
	}

	return signedToken, nil
}

// GenerateToken generates a new token.
func (a *authenticator) GenerateToken(record *api.Record) (string, error) {
	return a.generateToken(record)
}

func (a *authenticator) Authenticate(token string) (*api.Record, string, error) {
	var isExpired bool
	jwtToken, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return a.signingKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors == jwt.ValidationErrorExpired {
				isExpired = true
			} else {
				return nil, "", ecode.ErrInvalidToken
			}
		} else {
			return nil, "", ecode.ErrInvalidToken
		}
	}

	if jwtToken != nil {
		if claims, ok := jwtToken.Claims.(*CustomClaims); ok {
			// Check token issuer
			if claims.Issuer != a.issuer {
				return nil, "", ecode.ErrInvalidToken
			}

			var layout tokenLayout
			buf := bytes.NewBuffer(claims.Data)
			err = binary.Read(buf, binary.LittleEndian, &layout)
			if err != nil {
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

			if isExpired {
				// Check token refresh expiration time.
				refreshExpires := time.Unix(int64(layout.RefreshExpires), 0).UTC()
				if refreshExpires.Before(time.Now().Add(1 * time.Second)) {
					return nil, "", ecode.ErrTokenExpired
				}

				token, err = a.generateToken(record)
				return nil, token, nil
			} else {
				expires := time.Unix(claims.ExpiresAt, 0).UTC()
				record.Lifetime = time.Until(expires).String()
				if a.cacher != nil {
					if token != a.cacher.GetAuthToken(uid) {
						return nil, "", ecode.ErrInvalidToken
					}
				}
				return record, "", nil
			}
		}
		return nil, "", ecode.ErrInvalidToken
	}

	return nil, "", ecode.ErrMalformed
}
