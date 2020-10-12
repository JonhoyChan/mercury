package hash

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"mercury/x/config"
)

const defaultBCryptCost = 12

type HasherBCrypt struct {
	c HasherBCryptProvider
}

type HasherBCryptProvider interface {
	HasherBCrypt() *config.HasherBCryptConfig
}

func NewHasherBCrypt(c HasherBCryptProvider) *HasherBCrypt {
	return &HasherBCrypt{c: c}
}

func (b *HasherBCrypt) Hash(data []byte) ([]byte, error) {
	c := b.c.HasherBCrypt()
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
