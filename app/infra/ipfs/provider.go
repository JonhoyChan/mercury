package ipfs

import (
	"context"
	"io"
)

type Provider interface {
	ID() (string, error)

	Version() (string, string, error)

	Add(r io.Reader) (string, error)

	Cat(hash string) ([]byte, error)

	FilesWrite(ctx context.Context, dir, fileName string, r io.Reader) (string, error)

	FilesRead(ctx context.Context, dir, fileName string) ([]byte, error)
}
