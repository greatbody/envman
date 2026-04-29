package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/greatbody/envman/internal/config"
	"github.com/greatbody/envman/internal/profile"
)

func setupTestEnv(t *testing.T) (string, *config.Manager, *profile.Manager) {
	dir := t.TempDir()
	os.Setenv("ENVMAN_DIR", dir)
	cfgMgr = config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr = profile.NewManager(cfgMgr.ProfilesDir())
	unloadCmd.Flags().Set("all", "false")
	return dir, cfgMgr, profileMgr
}

func TestSplitComma(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", nil},
		{"single", "a", []string{"a"}},
		{"multiple", "a,b,c", []string{"a", "b", "c"}},
		{"spaces", " a , b , c ", []string{"a", "b", "c"}},
		{"empty_items", ",,,", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitComma(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected %s at %d, got %s", tt.expected[i], i, v)
				}
			}
		})
	}
}

func TestSplitString(t *testing.T) {
	result := splitString("a,b,c", ',')
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"  hello  ", "hello"},
		{"hello", "hello"},
		{"  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := trimSpace(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]string{
		"c": "3",
		"a": "1",
		"b": "2",
	}

	keys := sortedKeys(m)
	if len(keys) != 3 {
		t.Fatalf("expected 3, got %d", len(keys))
	}
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("unexpected order: %v", keys)
	}
}

func TestRootCommand(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("root command failed: %v", err)
	}
}

func TestListCommand(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Create("work")
	profileMgr.Create("personal")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"list"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected output from list command")
	}
}

func TestShowCommand(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Save("test", []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"show", "test"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("show command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected output from show command")
	}
}

func TestShowCommandMasked(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Save("test", []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"show", "test", "--masked"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("show command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected output from show command")
	}
}

func TestCreateCommand(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"create", "newprofile"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("create command failed: %v", err)
	}

	if !profileMgr.Exists("newprofile") {
		t.Error("profile should exist after create")
	}
}

func TestDeleteCommand(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Create("todelete")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"delete", "todelete"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("delete command failed: %v", err)
	}

	if profileMgr.Exists("todelete") {
		t.Error("profile should not exist after delete")
	}
}

func TestDeleteCommandDefault(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"delete", "default"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when deleting default profile")
	}
}

func TestCopyCommand(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Save("src", []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"copy", "src", "dst"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("copy command failed: %v", err)
	}

	if !profileMgr.Exists("dst") {
		t.Error("destination profile should exist")
	}
}

func TestDefaultCommand(t *testing.T) {
	_, cfgMgr, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Create("work")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"default", "work"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("default command failed: %v", err)
	}

	defaultName, _ := cfgMgr.GetDefaultProfile()
	if defaultName != "work" {
		t.Errorf("expected work, got %s", defaultName)
	}
}

func TestDiffCommand(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Save("a", []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY2", Value: "value2"},
	})
	profileMgr.Save("b", []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY3", Value: "value3"},
	})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"diff", "a", "b"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("diff command failed: %v", err)
	}
}

func TestDiffCommandIdentical(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Save("a", []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	})
	profileMgr.Save("b", []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"diff", "a", "b"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("diff command failed: %v", err)
	}
}

func TestEditCommand(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Create("test")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"edit", "test"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when EDITOR is not set")
	}
}

func TestLoadCommand(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Save("test", []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"load", "test"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("load command failed: %v", err)
	}
}

func TestLoadCommandNotExist(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"load", "nonexistent"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestUnloadCommand(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Save("test", []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	})

	os.Setenv("ENVMAN_LOADED_PROFILES", "test")
	defer os.Unsetenv("ENVMAN_LOADED_PROFILES")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"unload", "test"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unload command failed: %v", err)
	}
}

func TestUnloadCommandAll(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	os.Setenv("ENVMAN_LOADED_PROFILES", "base,work")
	defer os.Unsetenv("ENVMAN_LOADED_PROFILES")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"unload", "--all"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unload all command failed: %v", err)
	}
}

func TestUnloadCommandNoArgs(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Save("test", []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	})

	os.Setenv("ENVMAN_LOADED_PROFILES", "test")
	defer os.Unsetenv("ENVMAN_LOADED_PROFILES")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"unload"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unload command failed: %v", err)
	}
}

func TestUnloadCommandNoLoaded(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	os.Setenv("ENVMAN_LOADED_PROFILES", "")
	defer os.Unsetenv("ENVMAN_LOADED_PROFILES")

	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"unload"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when no profiles loaded")
	}
}

func TestExecute(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	rootCmd.SetArgs([]string{"list"})
	err := Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestShellGetLoadedProfiles(t *testing.T) {
	os.Setenv("ENVMAN_LOADED_PROFILES", "test1,test2")
	defer os.Unsetenv("ENVMAN_LOADED_PROFILES")

	result := shellGetLoadedProfiles()
	if result != "test1,test2" {
		t.Errorf("expected test1,test2, got %s", result)
	}
}

func TestShellGetLoadedProfilesEmpty(t *testing.T) {
	os.Unsetenv("ENVMAN_LOADED_PROFILES")

	result := shellGetLoadedProfiles()
	if result != "" {
		t.Errorf("expected empty, got %s", result)
	}
}

func TestShowCommandNotExist(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"show", "nonexistent"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestCreateCommandExists(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Create("existing")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"create", "existing"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for existing profile")
	}
}

func TestDeleteCommandNotExist(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"delete", "nonexistent"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestCopyCommandSrcNotExist(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"copy", "nonexistent", "dst"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

func TestCopyCommandDstExists(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Create("src")
	profileMgr.Create("dst")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"copy", "src", "dst"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when destination exists")
	}
}

func TestDefaultCommandNotExist(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"default", "nonexistent"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestDiffCommandANotExist(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Create("b")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"diff", "nonexistent", "b"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent profile a")
	}
}

func TestDiffCommandBNotExist(t *testing.T) {
	_, _, profileMgr := setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	profileMgr.Create("a")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"diff", "a", "nonexistent"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent profile b")
	}
}

func TestEditCommandNotExist(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"edit", "nonexistent"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestUnloadCommandNotExist(t *testing.T) {
	setupTestEnv(t)
	defer os.Unsetenv("ENVMAN_DIR")

	os.Setenv("ENVMAN_LOADED_PROFILES", "nonexistent")
	defer os.Unsetenv("ENVMAN_LOADED_PROFILES")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"unload", "nonexistent"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}
