package schema

import (
	"errors"
	"fmt"
	config "github.com/ipfs/go-ipfs-config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"mercury/x/ipfs"
	"os"
	"path/filepath"
	"runtime"
)

const (
	IdentityKeyLength = 4096
)

type nodeSchemaManager struct {
	dataPath    string
	identityKey []byte
	os          string
}

type NodeSchemaContext struct {
	DataPath    string
	IdentityKey []byte
	OS          string
}

func NewSchemaManager() (*nodeSchemaManager, error) {
	return NewCustomNodeSchemaManager(NodeSchemaContext{})
}

// NewCustomNodeSchemaManager allows a custom NodeSchemaContext to be provided
func NewCustomNodeSchemaManager(ctx NodeSchemaContext) (*nodeSchemaManager, error) {
	if len(ctx.DataPath) == 0 {
		path, err := MercuryPathTransform(defaultDataPath())
		if err != nil {
			return nil, fmt.Errorf("finding root path: %s", err.Error())
		}
		ctx.DataPath = path
	}
	if len(ctx.OS) == 0 {
		ctx.OS = runtime.GOOS
	}

	if len(ctx.IdentityKey) == 0 {
		identityKey, err := CreateIdentityKey()
		if err != nil {
			return nil, fmt.Errorf("generating identity: %s", err.Error())
		}
		ctx.IdentityKey = identityKey
	}

	return &nodeSchemaManager{
		dataPath:    ctx.DataPath,
		identityKey: ctx.IdentityKey,
		os:          ctx.OS,
	}, nil
}

func CreateIdentityKey() ([]byte, error) {
	identityKey, err := ipfs.IdentityKeyFromSeed([]byte("mercury-identity"), IdentityKeyLength)
	if err != nil {
		return nil, err
	}
	return identityKey, nil
}

func (m *nodeSchemaManager) IsInitialized() bool {
	return m.isConfigInitialized()
}

func (m *nodeSchemaManager) isConfigInitialized() bool {
	return fsrepo.IsInitialized(m.DataPath())
}

// DataPath returns the expected location of the data storage directory
func (m *nodeSchemaManager) DataPath() string { return m.dataPath }

// IdentityKey returns the identity key used by the schema
func (m *nodeSchemaManager) IdentityKey() []byte { return m.identityKey }

// Identity returns the struct representation of the []byte IdentityKey
func (m *nodeSchemaManager) Identity() (*config.Identity, error) {
	if len(m.identityKey) == 0 {
		// All public constructors set this value and should not occur during runtime
		return nil, errors.New("identity key is not generated")
	}
	identity, err := ipfs.IdentityFromKey(m.identityKey)
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (m *nodeSchemaManager) BuildSchemaDirectories() error {
	if err := os.MkdirAll(m.DataPathJoin("datastore"), os.ModePerm); err != nil {
		return err
	}
	if err := m.buildIPFSRootDirectories(); err != nil {
		return err
	}
	return nil
}

func (m *nodeSchemaManager) buildIPFSRootDirectories() error {
	if err := os.MkdirAll(m.DataPathJoin("root"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(m.DataPathJoin("root", "images"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(m.DataPathJoin("root", "images", "tiny"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(m.DataPathJoin("root", "images", "small"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(m.DataPathJoin("root", "images", "medium"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(m.DataPathJoin("root", "images", "large"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(m.DataPathJoin("root", "images", "original"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(m.DataPathJoin("root", "videos"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(m.DataPathJoin("root", "files"), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (m *nodeSchemaManager) DataPathJoin(pathArgs ...string) string {
	allPathArgs := append([]string{m.dataPath}, pathArgs...)
	return filepath.Join(allPathArgs...)
}

var BootstrapAddressesDefault = []string{
	"/ip4/192.168.120.88/tcp/4001/p2p/QmbRyQ8vJHnQ9KpfqA6KGBuBwqEFo69ttDzDKZ1Nbyrg8J",
}

func MustDefaultConfig() *config.Config {
	bootstrapPeers, err := config.ParseBootstrapPeers(BootstrapAddressesDefault)
	if err != nil {
		// BootstrapAddressesDefault are local and should never panic
		panic(err)
	}

	conf, err := config.Init(&dummyWriter{}, 4096)
	if err != nil {
		panic(err)
	}
	conf.Ipns.RecordLifetime = "168h"
	conf.Ipns.RepublishPeriod = "24h"
	conf.Discovery.MDNS.Enabled = false
	conf.Addresses = config.Addresses{
		Swarm: []string{
			"/ip4/0.0.0.0/tcp/4001",
			"/ip6/::/tcp/4001",
			"/ip4/0.0.0.0/tcp/9005/ws",
			"/ip6/::/tcp/9005/ws",
		},
		API:     []string{""},
		Gateway: []string{"/ip4/127.0.0.1/tcp/4002"},
	}
	conf.Bootstrap = config.BootstrapPeerStrings(bootstrapPeers)

	return conf
}

type dummyWriter struct{}

func (d *dummyWriter) Write(p []byte) (n int, err error) { return 0, nil }
