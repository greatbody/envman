package shell

import (
	"os"
	"testing"

	"github.com/greatbody/envman/internal/profile"
)

func TestExportVars(t *testing.T) {
	vars := []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY2", Value: "value with spaces"},
	}

	output := ExportVars(vars)

	if output != "export KEY1=\"value1\"\nexport KEY2=\"value with spaces\"" {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestExportVarsEmpty(t *testing.T) {
	output := ExportVars([]profile.EnvVar{})
	if output != "" {
		t.Errorf("expected empty output, got: %s", output)
	}
}

func TestUnsetVars(t *testing.T) {
	keys := []string{"KEY1", "KEY2"}

	output := UnsetVars(keys)

	if output != "unset KEY1\nunset KEY2" {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestUnsetVarsEmpty(t *testing.T) {
	output := UnsetVars([]string{})
	if output != "" {
		t.Errorf("expected empty output, got: %s", output)
	}
}

func TestExportWithTracking(t *testing.T) {
	vars := []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	}

	output := ExportWithTracking(vars, "work", "")

	expected := "export KEY1=\"value1\"\nexport ENVMAN_LOADED_PROFILES=\"work\""
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}

func TestExportWithTrackingExisting(t *testing.T) {
	vars := []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	}

	output := ExportWithTracking(vars, "work", "base")

	expected := "export KEY1=\"value1\"\nexport ENVMAN_LOADED_PROFILES=\"base,work\""
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}

func TestExportWithTrackingAlreadyLoaded(t *testing.T) {
	vars := []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	}

	output := ExportWithTracking(vars, "work", "work")

	expected := "export KEY1=\"value1\""
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}

func TestUnsetProfile(t *testing.T) {
	vars := []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY2", Value: "value2"},
	}

	output := UnsetProfile(vars, "work", "base,work")

	expected := "unset KEY1\nunset KEY2\nexport ENVMAN_LOADED_PROFILES=\"base\""
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}

func TestUnsetProfileLast(t *testing.T) {
	vars := []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	}

	output := UnsetProfile(vars, "work", "work")

	expected := "unset KEY1\nunset ENVMAN_LOADED_PROFILES"
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}

func TestUnsetAll(t *testing.T) {
	output := UnsetAll("base,work")

	expected := "# unloading profile: base\n# unloading profile: work\nunset ENVMAN_LOADED_PROFILES"
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}

func TestUnsetAllEmpty(t *testing.T) {
	output := UnsetAll("")

	expected := "unset ENVMAN_LOADED_PROFILES"
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}

func TestGetLoadedProfiles(t *testing.T) {
	os.Setenv(EnvVarLoadedProfiles, "base,work")
	defer os.Unsetenv(EnvVarLoadedProfiles)

	result := GetLoadedProfiles()
	if result != "base,work" {
		t.Errorf("expected base,work, got %s", result)
	}
}

func TestGetLoadedProfilesEmpty(t *testing.T) {
	os.Unsetenv(EnvVarLoadedProfiles)

	result := GetLoadedProfiles()
	if result != "" {
		t.Errorf("expected empty, got %s", result)
	}
}

func TestInitOutput(t *testing.T) {
	vars := []profile.EnvVar{
		{Key: "KEY1", Value: "value1"},
	}

	output := InitOutput(vars)

	expected := "# envman init\nexport KEY1=\"value1\""
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}
