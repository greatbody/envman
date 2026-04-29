package profile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type EnvVar struct {
	Key      string
	Value    string
	Comment  string
	Original string
}

type Profile struct {
	Name string
	Vars []EnvVar
}

type Manager struct {
	profilesDir string
}

func NewManager(profilesDir string) *Manager {
	return &Manager{profilesDir: profilesDir}
}

func (m *Manager) ProfilePath(name string) string {
	return filepath.Join(m.profilesDir, name+".env")
}

func (m *Manager) Exists(name string) bool {
	_, err := os.Stat(m.ProfilePath(name))
	return err == nil
}

func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.profilesDir)
	if err != nil {
		return nil, fmt.Errorf("reading profiles directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".env") {
			names = append(names, strings.TrimSuffix(name, ".env"))
		}
	}

	sort.Strings(names)
	return names, nil
}

func (m *Manager) Load(name string) (*Profile, error) {
	path := m.ProfilePath(name)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening profile %s: %w", name, err)
	}
	defer file.Close()

	profile := &Profile{Name: name}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		key, value, found := strings.Cut(trimmed, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		profile.Vars = append(profile.Vars, EnvVar{
			Key:      key,
			Value:    value,
			Original: line,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading profile %s: %w", name, err)
	}

	return profile, nil
}

func (m *Manager) Save(name string, vars []EnvVar) error {
	var lines []string
	for _, v := range vars {
		if v.Comment != "" {
			lines = append(lines, "# "+v.Comment)
		}
		lines = append(lines, fmt.Sprintf("%s=%s", v.Key, v.Value))
	}
	lines = append(lines, "")

	path := m.ProfilePath(name)
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

func (m *Manager) Create(name string) error {
	if m.Exists(name) {
		return fmt.Errorf("profile %s already exists", name)
	}
	return m.Save(name, []EnvVar{})
}

func (m *Manager) Delete(name string) error {
	if !m.Exists(name) {
		return fmt.Errorf("profile %s does not exist", name)
	}
	return os.Remove(m.ProfilePath(name))
}

func (m *Manager) Copy(src, dst string) error {
	if !m.Exists(src) {
		return fmt.Errorf("profile %s does not exist", src)
	}
	if m.Exists(dst) {
		return fmt.Errorf("profile %s already exists", dst)
	}

	data, err := os.ReadFile(m.ProfilePath(src))
	if err != nil {
		return fmt.Errorf("reading profile %s: %w", src, err)
	}

	return os.WriteFile(m.ProfilePath(dst), data, 0644)
}

func (m *Manager) GetVars(name string) (map[string]string, error) {
	profile, err := m.Load(name)
	if err != nil {
		return nil, err
	}

	vars := make(map[string]string)
	for _, v := range profile.Vars {
		vars[v.Key] = v.Value
	}
	return vars, nil
}

func (m *Manager) Diff(a, b string) (added, removed, changed map[string]string, err error) {
	varsA, err := m.GetVars(a)
	if err != nil {
		return nil, nil, nil, err
	}

	varsB, err := m.GetVars(b)
	if err != nil {
		return nil, nil, nil, err
	}

	added = make(map[string]string)
	removed = make(map[string]string)
	changed = make(map[string]string)

	for k, v := range varsB {
		if _, exists := varsA[k]; !exists {
			added[k] = v
		} else if varsA[k] != v {
			changed[k] = v
		}
	}

	for k, v := range varsA {
		if _, exists := varsB[k]; !exists {
			removed[k] = v
		}
	}

	return added, removed, changed, nil
}

func ParseEnvFile(content string) []EnvVar {
	var vars []EnvVar
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		key, value, found := strings.Cut(trimmed, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		vars = append(vars, EnvVar{
			Key:      key,
			Value:    value,
			Original: line,
		})
	}

	return vars
}
