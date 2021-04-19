package x

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseVersion(t *testing.T) {
	v1 := ParseVersion("0.1.0-dev")
	require.Equal(t, 256, v1)
	v2 := ParseVersion("0.2.0-dev")
	require.Equal(t, 512, v2)

	require.Equal(t, 1, VersionCompare(v2, v1))
}
