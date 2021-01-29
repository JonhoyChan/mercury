package lib

import (
	"context"
	"mercury/config"
	"mercury/x/ecode"
	"os"
	"path/filepath"
)

var (
	// ErrNoRepo is an error for  when a repo does not exist at a given path
	ErrNoRepo = ecode.NewError("no repo exists")
)

type Instance struct {
	cfg *config.Config

	cancel context.CancelFunc
	doneCh chan struct{}
}

func NewInstanceWithRepoPath(ctx context.Context, repoPath string) (inst *Instance, err error) {
	ctx, cancel := context.WithCancel(ctx)
	ok := false
	defer func() {
		if !ok {
			cancel()
		}
	}()

	if repoPath == "" {
		err = ecode.NewError("repo path is required")
		return
	}

	var cfg *config.Config
	if cfg, err = loadRepoConfig(repoPath); err != nil {
		return
	}

	if cfg == nil {
		// If at this point we don't have a configuration pointer one couldn't be
		// loaded from repoPath, and a configuration wasn't provided through Options,
		// so mercury needs to be set up
		err = ecode.NewError("no mercury repo found")
		return
	}

	inst = &Instance{
		cancel: cancel,
		doneCh: make(chan struct{}),
		cfg:    cfg,
	}

	ok = true
	return
}

func NewInstanceWithInfraUrl(ctx context.Context, infraUrl string) (inst *Instance, err error) {
	ctx, cancel := context.WithCancel(ctx)
	ok := false
	defer func() {
		if !ok {
			cancel()
		}
	}()

	if infraUrl == "" {
		err = ecode.NewError("infra url is required")
		return
	}

	var cfg *config.Config
	if cfg, err = loadRemoteConfig(ctx, infraUrl); err != nil {
		return
	}

	if cfg == nil {
		err = ecode.NewError("can not load config")
		return
	}

	inst = &Instance{
		cancel: cancel,
		doneCh: make(chan struct{}),
		cfg:    cfg,
	}

	ok = true
	return
}

// Config provides methods for manipulating Mercury configuration
func (inst *Instance) Config() *config.Config {
	if inst == nil {
		return nil
	}
	return inst.cfg
}

func (inst *Instance) Shutdown() <-chan error {
	errCh := make(chan error)
	go func() {
		<-inst.doneCh
		//errCh <- inst.doneErr
	}()
	inst.cancel()
	return errCh
}

func loadRepoConfig(repoPath string) (*config.Config, error) {
	path := filepath.Join(repoPath, "config.yaml")

	if _, e := os.Stat(path); os.IsNotExist(e) {
		return nil, nil
	}

	return config.ReadFromFile(path)
}

func loadRemoteConfig(ctx context.Context, url string) (*config.Config, error) {
	return config.ReadFromRemote(ctx, url)
}
