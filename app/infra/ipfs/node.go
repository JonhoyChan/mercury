package ipfs

import (
	"context"
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"io"
	"io/ioutil"
	"strings"
)

var (
	AvatarDir = "/avatar"
)

type Node struct {
	shell *shell.Shell
}

func NewNode(api string) (*Node, error) {
	a, err := ma.NewMultiaddr(strings.TrimSpace(api))
	if err != nil {
		return nil, err
	}

	_, host, err := manet.DialArgs(a)
	if err != nil {
		return nil, err
	}

	node := &Node{
		shell: shell.NewShell(host),
	}
	go node.up()
	return node, nil
}

func (n *Node) up() {
	defer func() {
		// TODO reconnection to ipfs node
		fmt.Println("ipfs node is down")
	}()

	up := n.shell.IsUp()
	for up {
		if !up {
			return
		}
		up = n.shell.IsUp()
	}
}

func (n *Node) ID() (string, error) {
	idOutput, err := n.shell.ID()
	if err != nil {
		return "", err
	}
	fmt.Println(idOutput.ID, idOutput.Addresses, idOutput.AgentVersion, idOutput.ProtocolVersion, idOutput.PublicKey)
	return idOutput.ID, nil
}

func (n *Node) Version() (string, string, error) {
	return n.shell.Version()
}

func (n *Node) Add(r io.Reader) (string, error) {
	return n.shell.Add(r)
}

func (n *Node) Cat(hash string) ([]byte, error) {
	reader, err := n.shell.Cat(hash)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(reader)
}

func (n *Node) FilesWrite(ctx context.Context, dir, fileName string, r io.Reader) (string, error) {
	opts := []shell.FilesOpt{
		shell.FilesWrite.Create(true),
		shell.FilesWrite.Parents(true),
		shell.FilesWrite.Hash("blake2b-256"),
	}

	path := dir + "/" + fileName
	err := n.shell.FilesWrite(ctx, path, r, opts...)

	stat, err := n.shell.FilesStat(ctx, path)
	if err != nil {
		return "", err
	}

	return stat.Hash, nil
}

func (n *Node) FilesRead(ctx context.Context, dir, fileName string) ([]byte, error) {
	path := dir + "/" + fileName
	reader, err := n.shell.FilesRead(ctx, path)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(reader)
}
