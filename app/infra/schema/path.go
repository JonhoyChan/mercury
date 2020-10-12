package schema

import (
	"github.com/mitchellh/go-homedir"
	"path/filepath"
	"runtime"
)

func defaultDataPath() (path string) {
	if runtime.GOOS == "darwin" {
		return "~/Library/Application Support"
	}
	return "~"
}

func GetRepoPath(repoPath string) (string, error) {
	ctx := NodeSchemaContext{}
	if repoPath != "" {
		ctx.DataPath = repoPath
	}
	nodeSchema, err := NewCustomNodeSchemaManager(ctx)
	if err != nil {
		return "", err
	}
	return nodeSchema.DataPath(), nil
}

func MercuryPathTransform(basePath string) (path string, err error) {
	path, err = homedir.Expand(filepath.Join(basePath, directoryName()))
	if err == nil {
		path = filepath.Clean(path)
	}
	return path, err
}

func directoryName() (directoryName string) {
	if runtime.GOOS == "linux" {
		directoryName = ".mercury"
	} else {
		directoryName = "mercury"
	}
	return directoryName
}
