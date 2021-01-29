package ipfs_test

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"mercury/app/infra/ipfs"
	"os"
	"path"
	"testing"
)

var node *ipfs.Node

func init() {
	var err error
	node, err = ipfs.NewNode("/ip4/127.0.0.1/tcp/5001")
	if err != nil {
		panic(err)
	}
}

func TestNode_Add(t *testing.T) {
	file, err := os.Open(path.Join("", "test-images/ultraman.jpg"))
	assert.NoError(t, err)
	defer file.Close()

	cid, err := node.Add(file)
	assert.NoError(t, err)

	assert.Equal(t, "QmaRQEiKhFrp63vivFsFA9DmgaDQJ3uCrWewxSYLniKtvR", cid)

	t.Logf("cid: %s", cid)
}

func TestNode_Cat(t *testing.T) {
	src, err := node.Cat("QmaRQEiKhFrp63vivFsFA9DmgaDQJ3uCrWewxSYLniKtvR")
	assert.NoError(t, err)

	out, err := os.Create(path.Join("", "test-images/QmaRQEiKhFrp63vivFsFA9DmgaDQJ3uCrWewxSYLniKtvR.jpeg"))
	assert.NoError(t, err)
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(src))
}

func TestNode_FilesWrite(t *testing.T) {
	file, err := os.Open(path.Join("", "test-images/ultraman.jpg"))
	assert.NoError(t, err)
	defer file.Close()

	cid, err := node.FilesWrite(context.Background(), ipfs.AvatarDir, "ultraman.jpg", file)
	assert.NoError(t, err)

	assert.Equal(t, "bafykbzacedzyrgcmcikxpwojdqz6tvggm2kqkgn4rf2cxqdsr64vroflo26qe", cid)

	t.Logf("cid: %s", cid)
}

func TestNode_FilesRead(t *testing.T) {
	src, err := node.FilesRead(context.Background(), ipfs.AvatarDir, "ultraman.jpg")
	assert.NoError(t, err)

	out, err := os.Create(path.Join("", "test-images/bafykbzacedzyrgcmcikxpwojdqz6tvggm2kqkgn4rf2cxqdsr64vroflo26qe.jpeg"))
	assert.NoError(t, err)
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(src))
}
