package service

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/golang-lru"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"io/ioutil"
	"mercury/app/infra/config"
	"mercury/app/infra/model"
	"mercury/app/infra/schema"
	"mercury/version"
	"mercury/x"
	"mercury/x/ecode"
	"mercury/x/ipfs"
	"mercury/x/log"
	"mercury/x/secretboxer"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var cache, _ = lru.New(3)

type Service struct {
	c        config.Provider
	boxer    *secretboxer.PassphraseBoxer
	ipfsNode *core.IpfsNode
	repoRoot string
}

func NewService(c config.Provider) (*Service, error) {
	s := &Service{
		c:     c,
		boxer: secretboxer.NewPassphraseBoxer(version.String, secretboxer.EncodingTypeStd),
	}

	if err := s.load(); err != nil {
		return nil, err
	}
	if err := s.setupIPFS(); err != nil {
		return nil, err
	}
	return s, nil
}

func doInit(repoRoot string) error {
	nodeSchema, err := schema.NewCustomNodeSchemaManager(schema.NodeSchemaContext{
		DataPath: repoRoot,
	})
	if err != nil {
		return err
	}

	if !nodeSchema.IsInitialized() {
		if err := nodeSchema.BuildSchemaDirectories(); err != nil {
			return err
		}

		if err := writeSwarmKey(repoRoot); err != nil {
			return err
		}

		conf := schema.MustDefaultConfig()
		if err := fsrepo.Init(repoRoot, conf); err != nil {
			return err
		}
	}

	return nil
}

var swarmKey = `
/key/swarm/psk/1.0.0/
/base16/
64d5790c69567320e975b7e6e594b6c7f7d297746d0401070ab52739fdcf2fd8
`

func writeSwarmKey(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(dir, "swarm.key"))
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("%s is not writeable by the current user", dir)
		}
		return fmt.Errorf("unexpected error while creating swarm.key of repo root: %s", err)
	}
	f.WriteString(strings.TrimSpace(swarmKey))
	_ = f.Close()
	return nil

}

func (s *Service) setupIPFS() error {
	if s.ipfsNode == nil && s.repoRoot == "" {
		ipfs.InstallDatabasePlugins()

		repoPath, err := schema.GetRepoPath(s.c.RepoPath())
		if err != nil {
			return err
		}

		repoLockFile := filepath.Join(repoPath, fsrepo.LockFile)
		_ = os.Remove(repoLockFile)

		if err := doInit(repoPath); err != nil {
			return err
		}

		r, err := fsrepo.Open(repoPath)
		if err != nil {
			log.Error("failed to open repo:", err)
			return err
		}

		node, err := core.NewNode(context.Background(), ipfs.PrepareIPFSConfig(r))
		if err != nil {
			log.Error("failed to create new ipfs node:", err)
			return err
		}
		node.IsDaemon = true

		s.ipfsNode = node
		s.repoRoot = repoPath
		if s.ipfsNode.PNetFingerprint != nil {
			log.Info("Swarm is limited to private network of peers with the swarm key", "fingerprint", x.Sprintf("%x", s.ipfsNode.PNetFingerprint))
		}
	}

	return nil
}

func (s *Service) load() error {
	configPath := s.c.ConfigPath()
	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		filename := file.Name()
		v := viper.New()
		p := x.Sprintf("%s/%s", s.c.ConfigPath(), filename)
		v.SetConfigFile(p)
		if err := v.ReadInConfig(); err != nil {
			return err
		}

		_ = cache.Add(strings.TrimSuffix(filename, ".yml"), v)
		go s.watch(v)
	}
	return nil
}

func (s *Service) watch(v *viper.Viper) {
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Info("[watch] The configuration file has changed", "filename", e.Name)
	})
}

func (s *Service) LoadConfig(name string) (string, error) {
	var allSettings map[string]interface{}
	if v, found := cache.Get(name); found {
		allSettings = v.(*viper.Viper).AllSettings()
	} else {
		filename := name
		if !strings.HasSuffix(filename, ".yml") {
			filename += ".yml"
		}
		p := x.Sprintf("%s/%s", s.c.ConfigPath(), filename)
		f, err := os.Open(p)
		if err != nil {
			return "", err
		}

		v := viper.New()
		v.SetConfigFile(p)
		if err := v.ReadInConfig(); err != nil {
			return "", err
		}
		_ = cache.Add(f.Name(), v)

		allSettings = v.AllSettings()
	}

	b, err := jsoniter.Marshal(allSettings)
	if err != nil {
		return "", err
	}

	ciphertext, err := s.boxer.Seal(b)
	if err != nil {
		return "", err
	}

	return ciphertext, nil
}

func (s *Service) AddImages(imageData []byte, filename string) (*model.ProfileImage, error) {
	return s.resizeImage(imageData, filename, 120, 120)
}

func (s *Service) AddVideos(videoData []byte, filename string) (*model.ProfileVideo, error) {
	p := path.Join(s.repoRoot, "root", "videos")
	original, err := s.addVideo(videoData, path.Join(p, filename))
	if err != nil {
		return nil, err
	}
	return &model.ProfileVideo{Original: original}, nil
}

func (s *Service) addVideo(data []byte, path string) (string, error) {
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	out.Write(data)
	out.Close()
	return ipfs.AddFile(s.ipfsNode, path)
}

func (s *Service) CatFile(hash string) ([]byte, error) {
	data, err := ipfs.Cat(s.ipfsNode, hash, 30*time.Second)
	if err != nil {
		return nil, ecode.ErrDataDoesNotExist
	}
	return data, nil
}
