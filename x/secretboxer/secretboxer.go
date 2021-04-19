package secretboxer

import (
	"fmt"
)

const (
	WrapTypePassphrase = "passphrase"
)

const (
	EncodingTypeStd = "std"
	EncodingTypeURL = "url"
)

type SecretBoxer interface {
	// Encrypt the byte fragment and return a base64 string.
	Encrypt(in []byte) (string, error)
	// Decrypt the encrypted base64 string and return the decrypted byte fragment.
	Decrypt(in string) ([]byte, error)
	// Return the wrap type of SecretBoxer(e.g, passphrase).
	WrapType() string
	// Return the encoding type of SecretBoxer (e.g, std, url).
	EncodingType() string
}

func SecretBoxerForType(wrapType, passphrase, encodingType string) (SecretBoxer, error) {
	switch wrapType {
	case WrapTypePassphrase:
		return NewPassphraseBoxer(passphrase, encodingType), nil
	default:
		return nil, fmt.Errorf("unknown secret boxer: %s", wrapType)
	}
}
