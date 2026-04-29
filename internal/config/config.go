package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultProfileName = "default"
	EnvManDir          = ".envman"
	ProfilesDir        = "profiles"
	ConfigFile         = "config.json"
	EnvVarLoadedProfiles = "ENVMAN_LOADED_PROFILES"
)

type Config struct {
	DefaultProfile string `json:"default_profile"`
}

type Manager struct {
	baseDir string
}

func NewManager() *Manager {
	if envDir := os.Getenv("ENVMAN_DIR"); envDir != "" {
		return &Manager{baseDir: envDir}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &Manager{
		baseDir: filepath.Join(home, EnvManDir),
	}
}

func NewManagerWithBaseDir(dir string) *Manager {
	return &Manager{baseDir: dir}
}

func (m *Manager) BaseDir() string {
	return m.baseDir
}

func (m *Manager) ProfilesDir() string {
	return filepath.Join(m.baseDir, ProfilesDir)
}

func (m *Manager) ConfigPath() string {
	return filepath.Join(m.baseDir, ConfigFile)
}

func (m *Manager) Init() error {
	if err := os.MkdirAll(m.ProfilesDir(), 0755); err != nil {
		return fmt.Errorf("creating profiles directory: %w", err)
	}

	if _, err := os.Stat(m.ConfigPath()); os.IsNotExist(err) {
		cfg := Config{DefaultProfile: DefaultProfileName}
		if err := m.Save(&cfg); err != nil {
			return err
		}
	}

	defaultProfile := filepath.Join(m.ProfilesDir(), DefaultProfileName+".env")
	if _, err := os.Stat(defaultProfile); os.IsNotExist(err) {
		if err := os.WriteFile(defaultProfile, []byte("# Default environment variables\n"), 0644); err != nil {
			return fmt.Errorf("creating default profile: %w", err)
		}
	}

	return nil
}

func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

func (m *Manager) Save(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(m.ConfigPath(), data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

func (m *Manager) GetDefaultProfile() (string, error) {
	cfg, err := m.Load()
	if err != nil {
		return DefaultProfileName, nil
	}
	return cfg.DefaultProfile, nil
}

func (m *Manager) SetDefaultProfile(name string) error {
	cfg, err := m.Load()
	if err != nil {
		cfg = &Config{DefaultProfile: DefaultProfileName}
	}
	cfg.DefaultProfile = name
	return m.Save(cfg)
}
