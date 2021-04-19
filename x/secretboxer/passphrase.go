package secretboxer

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	saltLength  = 16
	nonceLength = 24
	keyLength   = 32
)

func deriveKey(passphrase string, salt []byte) [keyLength]byte {
	secretKeyBytes := argon2.IDKey([]byte(passphrase), salt, 4, 64*1024, 4, 32)
	var secretKey [keyLength]byte
	copy(secretKey[:], secretKeyBytes)
	return secretKey
}

type PassphraseBoxer struct {
	passphrase   string
	encodingType string
}

func NewPassphraseBoxer(passphrase, encodingType string) *PassphraseBoxer {
	if encodingType == "" {
		encodingType = EncodingTypeStd
	}
	return &PassphraseBoxer{
		passphrase:   passphrase,
		encodingType: encodingType,
	}
}

// Return the wrap type of PassphraseBoxer.
func (b *PassphraseBoxer) WrapType() string {
	return WrapTypePassphrase
}

// Return the encoding type of PassphraseBoxer.
func (b *PassphraseBoxer) EncodingType() string {
	return b.encodingType
}

// Encrypt the byte fragment and return a base64 string.
func (b *PassphraseBoxer) Encrypt(in []byte) (string, error) {
	var nonce [nonceLength]byte
	// generate a random nonce
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return "", err
	}

	salt := make([]byte, saltLength)
	// Generate a random salt
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Based on passphrase, salt, derive a secret key
	secretKey := deriveKey(b.passphrase, salt)
	prefix := append(salt, nonce[:]...)
	cipherText := secretbox.Seal(prefix, in, &nonce, &secretKey)

	switch b.encodingType {
	case EncodingTypeURL:
		return base64.RawURLEncoding.EncodeToString(cipherText), nil
	default:
		return base64.RawStdEncoding.EncodeToString(cipherText), nil
	}
}

// Decrypt the encrypted base64 string and return the decrypted byte fragment.
func (b *PassphraseBoxer) Decrypt(in string) ([]byte, error) {
	var (
		buf []byte
		err error
	)
	switch b.encodingType {
	case EncodingTypeURL:
		buf, err = base64.RawURLEncoding.DecodeString(in)
	default:
		buf, err = base64.RawStdEncoding.DecodeString(in)
	}
	if err != nil {
		return []byte{}, err
	}

	salt := make([]byte, saltLength)
	copy(salt, buf[:saltLength])
	var nonce [nonceLength]byte
	copy(nonce[:], buf[saltLength:nonceLength+saltLength])

	// Based on passphrase, salt, derive a secret key
	secretKey := deriveKey(b.passphrase, salt)

	decrypted, ok := secretbox.Open(nil, buf[nonceLength+saltLength:], &nonce, &secretKey)
	if !ok {
		return []byte{}, fmt.Errorf("failed to decrypt")
	}
	return decrypted, nil
}
