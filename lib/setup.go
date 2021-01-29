package lib

import (
	"fmt"
	"mercury/config"
	"os"
	"path/filepath"
)

func RepoExists(path string) error {
	// for now this just checks for an existing config file
	_, err := os.Stat(filepath.Join(path, "config.yaml"))
	if !os.IsNotExist(err) {
		return nil
	}
	return ErrNoRepo
}

// SetupParams encapsulates arguments for Setup
type SetupParams struct {
	// a configuration is required. defaults to config.DefaultConfig()
	Config *config.Config
	// where to initialize mercury repository
	RepoPath string
}

func Setup(p SetupParams) error {
	if err := RepoExists(p.RepoPath); err == nil {
		return fmt.Errorf("repo already initialized")
	}

	cfg := p.Config
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	if err := setup(p.RepoPath, cfg); err != nil {
		return err
	}
	return nil
}

func setup(repoPath string, cfg *config.Config) error {
	if err := os.MkdirAll(repoPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating mercury repo directory: %s, path: %s", err.Error(), repoPath)
	}

	cfgPath := filepath.Join(repoPath, "config.yaml")

	if err := cfg.WriteToFile(cfgPath); err != nil {
		return fmt.Errorf("error writing config: %s", err.Error())
	}
	return nil
}
