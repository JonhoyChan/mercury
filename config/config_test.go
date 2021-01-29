package config

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReadFromRemote(t *testing.T) {
	_, err := ReadFromRemote(context.TODO(), "http://localhost:9600/infra/v1")
	require.NoError(t, err)
}
