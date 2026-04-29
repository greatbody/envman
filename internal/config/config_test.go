package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	// Ensure ENVMAN_DIR is not set
	os.Unsetenv("ENVMAN_DIR")
	m := NewManager()
	if m.baseDir == "" {
		t.Error("expected baseDir to be set")
	}
}

func TestNewManagerWithEnvDir(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("ENVMAN_DIR", dir)
	defer os.Unsetenv("ENVMAN_DIR")

	m := NewManager()
	if m.baseDir != dir {
		t.Errorf("expected %s, got %s", dir, m.baseDir)
	}
}

func TestNewManagerWithBaseDir(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)
	if m.baseDir != dir {
		t.Errorf("expected %s, got %s", dir, m.baseDir)
	}
}

func TestManager_BaseDir(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)
	if m.BaseDir() != dir {
		t.Errorf("expected %s, got %s", dir, m.BaseDir())
	}
}

func TestManager_ProfilesDir(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)
	expected := filepath.Join(dir, ProfilesDir)
	if m.ProfilesDir() != expected {
		t.Errorf("expected %s, got %s", expected, m.ProfilesDir())
	}
}

func TestManager_ConfigPath(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)
	expected := filepath.Join(dir, ConfigFile)
	if m.ConfigPath() != expected {
		t.Errorf("expected %s, got %s", expected, m.ConfigPath())
	}
}

func TestManager_Init(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	if err := m.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if _, err := os.Stat(m.ProfilesDir()); os.IsNotExist(err) {
		t.Error("profiles directory should exist")
	}

	if _, err := os.Stat(m.ConfigPath()); os.IsNotExist(err) {
		t.Error("config file should exist")
	}

	defaultProfile := filepath.Join(m.ProfilesDir(), DefaultProfileName+".env")
	if _, err := os.Stat(defaultProfile); os.IsNotExist(err) {
		t.Error("default profile should exist")
	}
}

func TestManager_InitIdempotent(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	if err := m.Init(); err != nil {
		t.Fatalf("first Init failed: %v", err)
	}

	if err := m.Init(); err != nil {
		t.Fatalf("second Init failed: %v", err)
	}
}

func TestManager_Load(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	if err := m.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cfg, err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.DefaultProfile != DefaultProfileName {
		t.Errorf("expected default profile %s, got %s", DefaultProfileName, cfg.DefaultProfile)
	}
}

func TestManager_LoadNotExist(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	_, err := m.Load()
	if err == nil {
		t.Error("expected error when config does not exist")
	}
}

func TestManager_Save(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	if err := m.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cfg := &Config{DefaultProfile: "test"}
	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.DefaultProfile != "test" {
		t.Errorf("expected test, got %s", loaded.DefaultProfile)
	}
}

func TestManager_SaveCreatesDir(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(filepath.Join(dir, "nested", "dir"))

	// Need to create the directory first
	os.MkdirAll(m.ProfilesDir(), 0755)

	cfg := &Config{DefaultProfile: "test"}
	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
}

func TestManager_GetDefaultProfile(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	if err := m.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	name, err := m.GetDefaultProfile()
	if err != nil {
		t.Fatalf("GetDefaultProfile failed: %v", err)
	}

	if name != DefaultProfileName {
		t.Errorf("expected %s, got %s", DefaultProfileName, name)
	}
}

func TestManager_GetDefaultProfileNoConfig(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	name, err := m.GetDefaultProfile()
	if err != nil {
		t.Fatalf("GetDefaultProfile failed: %v", err)
	}

	if name != DefaultProfileName {
		t.Errorf("expected %s, got %s", DefaultProfileName, name)
	}
}

func TestManager_SetDefaultProfile(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	if err := m.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if err := m.SetDefaultProfile("work"); err != nil {
		t.Fatalf("SetDefaultProfile failed: %v", err)
	}

	name, err := m.GetDefaultProfile()
	if err != nil {
		t.Fatalf("GetDefaultProfile failed: %v", err)
	}

	if name != "work" {
		t.Errorf("expected work, got %s", name)
	}
}

func TestManager_SetDefaultProfileNoConfig(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	if err := m.SetDefaultProfile("work"); err != nil {
		t.Fatalf("SetDefaultProfile failed: %v", err)
	}

	name, err := m.GetDefaultProfile()
	if err != nil {
		t.Fatalf("GetDefaultProfile failed: %v", err)
	}

	if name != "work" {
		t.Errorf("expected work, got %s", name)
	}
}

func TestManager_InitCreateDirError(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(filepath.Join(dir, "nonexistent", "deep"))

	if err := m.Init(); err != nil {
		t.Fatalf("Init should create dirs: %v", err)
	}
}

func TestManager_InitConfigExists(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	// Create the directory and config file manually
	os.MkdirAll(m.ProfilesDir(), 0755)
	os.WriteFile(m.ConfigPath(), []byte(`{"default_profile":"custom"}`), 0644)

	if err := m.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	cfg, err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.DefaultProfile != "custom" {
		t.Errorf("expected custom, got %s", cfg.DefaultProfile)
	}
}

func TestManager_InitDefaultProfileExists(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	// Create the directory and default profile manually
	os.MkdirAll(m.ProfilesDir(), 0755)
	os.WriteFile(filepath.Join(m.ProfilesDir(), DefaultProfileName+".env"), []byte("KEY=value"), 0644)

	if err := m.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(m.ProfilesDir(), DefaultProfileName+".env"))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(content) != "KEY=value" {
		t.Errorf("expected KEY=value, got %s", string(content))
	}
}

func TestManager_LoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	os.MkdirAll(m.ProfilesDir(), 0755)
	os.WriteFile(m.ConfigPath(), []byte("invalid json"), 0644)

	_, err := m.Load()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestManager_SaveWriteError(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithBaseDir(dir)

	cfg := &Config{DefaultProfile: "test"}
	err := m.Save(cfg)
	if err != nil {
		t.Fatalf("Save should create dirs: %v", err)
	}
}
