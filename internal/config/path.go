package config

import (
	"os"
	"path/filepath"
)

type Paths struct {
	Root       string
	ConfigDir  string
	ConfigFile string
	DataDir    string
	Database   string
	LogDir     string
	Resources  string
	Libs       string
}

func ResolvePortablePaths() (Paths, error) {
	exe, err := os.Executable()
	if err != nil {
		return Paths{}, err
	}

	root := filepath.Dir(exe)
	configDir := filepath.Join(root, "config")

	dataDir := filepath.Join(root, "data")

	return Paths{
		Root:       root,
		ConfigDir:  configDir,
		ConfigFile: filepath.Join(configDir, "config.json"),
		DataDir:    dataDir,
		Database:   filepath.Join(dataDir, "keva.db"),
		LogDir:     filepath.Join(root, "logs"),
		Resources:  filepath.Join(root, "resources"),
		Libs:       filepath.Join(root, "libs"),
	}, nil
}

func EnsurePortableDirs(paths Paths) error {
	for _, dir := range []string{paths.ConfigDir, paths.DataDir, paths.LogDir, paths.Resources, paths.Libs} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
