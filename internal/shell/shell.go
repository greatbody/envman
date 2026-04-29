package shell

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/greatbody/envman/internal/profile"
)

const EnvVarLoadedProfiles = "ENVMAN_LOADED_PROFILES"

func ExportVars(vars []profile.EnvVar) string {
	var lines []string
	for _, v := range vars {
		lines = append(lines, fmt.Sprintf("export %s=%q", v.Key, v.Value))
	}
	return strings.Join(lines, "\n")
}

func UnsetVars(keys []string) string {
	var lines []string
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("unset %s", k))
	}
	return strings.Join(lines, "\n")
}

func ExportWithTracking(vars []profile.EnvVar, profileName string, existingLoaded string) string {
	var lines []string

	for _, v := range vars {
		lines = append(lines, fmt.Sprintf("export %s=%q", v.Key, v.Value))
	}

	var loaded []string
	if existingLoaded != "" {
		loaded = strings.Split(existingLoaded, ",")
		for _, l := range loaded {
			if l == profileName {
				return strings.Join(lines, "\n")
			}
		}
	}
	loaded = append(loaded, profileName)

	lines = append(lines, fmt.Sprintf("export %s=%q", EnvVarLoadedProfiles, strings.Join(loaded, ",")))

	return strings.Join(lines, "\n")
}

func UnsetProfile(vars []profile.EnvVar, profileName string, loadedProfiles string) string {
	var lines []string

	loaded := strings.Split(loadedProfiles, ",")
	var remaining []string
	for _, l := range loaded {
		if l != profileName {
			remaining = append(remaining, l)
		}
	}

	keys := make([]string, len(vars))
	for i, v := range vars {
		keys[i] = v.Key
	}
	sort.Strings(keys)

	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("unset %s", k))
	}

	if len(remaining) > 0 {
		lines = append(lines, fmt.Sprintf("export %s=%q", EnvVarLoadedProfiles, strings.Join(remaining, ",")))
	} else {
		lines = append(lines, fmt.Sprintf("unset %s", EnvVarLoadedProfiles))
	}

	return strings.Join(lines, "\n")
}

func UnsetAll(loadedProfiles string) string {
	loaded := strings.Split(loadedProfiles, ",")
	var lines []string

	for _, name := range loaded {
		if name == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("# unloading profile: %s", name))
	}

	lines = append(lines, fmt.Sprintf("unset %s", EnvVarLoadedProfiles))

	return strings.Join(lines, "\n")
}

func GetLoadedProfiles() string {
	return os.Getenv(EnvVarLoadedProfiles)
}

func InitOutput(defaultVars []profile.EnvVar) string {
	var lines []string
	lines = append(lines, "# envman init")
	lines = append(lines, ExportVars(defaultVars))
	return strings.Join(lines, "\n")
}
