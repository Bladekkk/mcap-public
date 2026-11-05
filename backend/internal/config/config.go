package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	AppName         string
	AppVersion      string
	DataDir         string
	CacheDir        string
	ServerPort      string
	RCONEnabled     bool
	RCONPort        int
	RCONPassword    string
	RCONTimeoutSec  int
	ModrinthToken   string
	CurseForgeKey   string
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	appDir := filepath.Join(home, ".mcap")
	return &Config{
		AppName:         "MCAP",
		AppVersion:      "0.1.0",
		DataDir:         filepath.Join(appDir, "data"),
		CacheDir:        filepath.Join(appDir, "cache"),
		ServerPort:      "8080",
		RCONEnabled:     true,
		RCONPort:        25575,
		RCONPassword:    "change_me_in_production",
		RCONTimeoutSec:  5,
	}
}

func LoadConfig() *Config {
	cfg := DefaultConfig()
	_ = os.MkdirAll(cfg.DataDir, 0755)
	_ = os.MkdirAll(cfg.CacheDir, 0755)
	_ = os.MkdirAll(filepath.Join(cfg.DataDir, "servers"), 0755)
	_ = os.MkdirAll(filepath.Join(cfg.DataDir, "jdk"), 0755)
	_ = os.MkdirAll(filepath.Join(cfg.CacheDir, "jars"), 0755)
	return cfg
}

func GetServerDataDir(baseDir, serverID string) string {
	return filepath.Join(baseDir, "servers", serverID)
}

func GetJDKCacheDir(baseDir string) string {
	return filepath.Join(baseDir, "jdk")
}

func GetJARCacheDir(baseDir string) string {
	return filepath.Join(baseDir, "cache", "jars")
}

func (c *Config) Validate() error {
	if c.RCONPort < 1 || c.RCONPort > 65535 {
		return fmt.Errorf("invalid RCON port: %d", c.RCONPort)
	}
	return nil
}
