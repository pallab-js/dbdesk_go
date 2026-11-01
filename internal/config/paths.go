package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const appDirName = "dbdesk"

type Paths struct {
	DataDir  string
	DBPath   string
	DemoDir  string
	DemoPath string
	LogPath  string
}

func Ensure() (Paths, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, fmt.Errorf("find home directory: %w", err)
		}
		base = filepath.Join(home, ".local", "share")
	}
	dataDir := filepath.Join(base, appDirName)
	p := Paths{
		DataDir:  dataDir,
		DBPath:   filepath.Join(dataDir, "app.db"),
		DemoDir:  filepath.Join(dataDir, "demo"),
		DemoPath: filepath.Join(dataDir, "demo", "ecommerce.db"),
		LogPath:  filepath.Join(dataDir, "logs", "app.log"),
	}
	if err := os.MkdirAll(p.DemoDir, 0o755); err != nil {
		return Paths{}, fmt.Errorf("create app data directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(p.LogPath), 0o755); err != nil {
		return Paths{}, fmt.Errorf("create log directory: %w", err)
	}
	return p, nil
}
