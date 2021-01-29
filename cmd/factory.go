package cmd

import (
	"context"
	"github.com/mitchellh/go-homedir"
	"mercury/config"
	"mercury/lib"
	"mercury/x/ecode"
	"mercury/x/log"
	"os"
	"path/filepath"
)

// Factory is an interface for providing required structures to cobra commands
// It's main implementation is QriOptions
type Factory interface {
	Init() error

	Instance() *lib.Instance
	// path to mercury data directory
	RepoPath() string
	Config() (*config.Config, error)

	InfraServer() (*lib.InfraServer, error)
	JobServer() (*lib.JobServer, error)
	CometServer() (*lib.CometServer, error)
	LogicServer() (*lib.LogicServer, error)
}

// StandardRepoPath returns mercury paths based on the MERCURY_PATH environment
// variable falling back to the default: $HOME/.mercury
func StandardRepoPath() string {
	repoPath := os.Getenv("MERCURY_PATH")
	if repoPath == "" {
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		repoPath = filepath.Join(home, ".mercury")
	}

	return repoPath
}

type Options struct {
	doneCh chan struct{}

	repoPath string
	infraUrl string

	log  log.Logger
	inst *lib.Instance

	infra bool
}

// NewOptions creates an options object
func NewOptions(repoPath string) *Options {
	return &Options{
		doneCh:   make(chan struct{}),
		repoPath: repoPath,
	}
}

func (o *Options) Init() (err error) {
	repoErr := lib.RepoExists(o.repoPath)
	if repoErr != nil {
		return ecode.NewError("no mercury repo exists")
	}

	ctx := context.Background()
	if o.infra {
		o.inst, err = lib.NewInstanceWithRepoPath(ctx, o.repoPath)
		if err != nil {
			return
		}
	} else {
		o.inst, err = lib.NewInstanceWithInfraUrl(ctx, o.infraUrl)
	}
	if err != nil {
		return err
	}

	lvl, _ := log.LvlFromString(o.inst.Config().LogLevel)
	log.Root().SetHandler(log.LvlFilterHandler(lvl, log.StreamHandler(os.Stdout, log.TerminalFormat(true))))

	o.log = log.New()
	return
}

// Instance returns the instance this options is using
func (o *Options) Instance() *lib.Instance {
	if err := o.Init(); err != nil {
		return nil
	}
	return o.inst
}

// RepoPath returns the path to the mercury data directory
func (o *Options) RepoPath() string {
	return o.repoPath
}

// Config returns from internal state
func (o *Options) Config() (*config.Config, error) {
	if err := o.Init(); err != nil {
		return nil, err
	}
	return o.inst.Config(), nil
}

func (o *Options) InfraServer() (*lib.InfraServer, error) {
	o.infra = true
	if err := o.Init(); err != nil {
		return nil, err
	}
	return lib.NewInfraServer(o.inst, o.log.New("lib", "infra")), nil
}

func (o *Options) JobServer() (*lib.JobServer, error) {
	if err := o.Init(); err != nil {
		return nil, err
	}
	return lib.NewJobServer(o.inst, o.log.New("lib", "job")), nil
}

func (o *Options) CometServer() (*lib.CometServer, error) {
	if err := o.Init(); err != nil {
		return nil, err
	}
	return lib.NewCometServer(o.inst, o.log.New("lib", "comet")), nil
}

func (o *Options) LogicServer() (*lib.LogicServer, error) {
	if err := o.Init(); err != nil {
		return nil, err
	}
	return lib.NewLogicServer(o.inst, o.log.New("lib", "logic")), nil
}

func (o *Options) Shutdown() <-chan error {
	if o.inst == nil {
		done := make(chan error)
		go func() { done <- nil }()
		return done
	}
	return o.inst.Shutdown()
}
