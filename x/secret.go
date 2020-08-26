package x

import (
	"crypto/rand"
	"math/big"
)

var rander = rand.Reader // random function

// RuneSequence returns a random sequence using the defined allowed runes.
func RuneSequence(l int, allowedRunes []rune) (seq []rune, err error) {
	c := big.NewInt(int64(len(allowedRunes)))
	seq = make([]rune, l)
	for i := 0; i < l; i++ {
		r, err := rand.Int(rander, c)
		if err != nil {
			return seq, err
		}
		seq[i] = allowedRunes[r.Uint64()]
	}
	return seq, nil
}

var secretCharSet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_-.~")

func GenerateSecret(length int) ([]byte, error) {
	secret, err := RuneSequence(length, secretCharSet)
	if err != nil {
		return nil, err
	}
	return []byte(string(secret)), nil
}
