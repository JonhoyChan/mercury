package hash

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const defaultBCryptCost = 12

type HasherBCrypt struct {
	c HasherConfigProvider
}

func NewHasherBCrypt(c HasherConfigProvider) *HasherBCrypt {
	return &HasherBCrypt{c: c}
}

func (b *HasherBCrypt) Hash(data []byte) ([]byte, error) {
	c := b.c.Hasher().BCrypt
	cost := c.Cost
	if cost == 0 {
		cost = defaultBCryptCost
	}
	s, err := bcrypt.GenerateFromPassword(data, cost)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return s, nil
}

func (b *HasherBCrypt) Compare(hash, data []byte) error {
	if err := bcrypt.CompareHashAndPassword(hash, data); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
