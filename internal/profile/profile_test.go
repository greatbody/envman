package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	if m.profilesDir != dir {
		t.Errorf("expected %s, got %s", dir, m.profilesDir)
	}
}

func TestManager_ProfilePath(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	expected := filepath.Join(dir, "test.env")
	if m.ProfilePath("test") != expected {
		t.Errorf("expected %s, got %s", expected, m.ProfilePath("test"))
	}
}

func TestManager_Exists(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	if m.Exists("nonexistent") {
		t.Error("expected false for nonexistent profile")
	}

	os.WriteFile(filepath.Join(dir, "test.env"), []byte("KEY=value"), 0644)

	if !m.Exists("test") {
		t.Error("expected true for existing profile")
	}
}

func TestManager_List(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	names, err := m.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(names))
	}

	os.WriteFile(filepath.Join(dir, "b.env"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "a.env"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte(""), 0644)

	names, err = m.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(names))
	}
	if names[0] != "a" || names[1] != "b" {
		t.Errorf("expected [a b], got %v", names)
	}
}

func TestManager_Load(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	content := `# Comment
KEY1=value1
KEY2="quoted"
KEY3='single'
EMPTY=
`
	os.WriteFile(filepath.Join(dir, "test.env"), []byte(content), 0644)

	p, err := m.Load("test")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if p.Name != "test" {
		t.Errorf("expected test, got %s", p.Name)
	}

	if len(p.Vars) != 4 {
		t.Fatalf("expected 4 vars, got %d", len(p.Vars))
	}

	if p.Vars[0].Key != "KEY1" || p.Vars[0].Value != "value1" {
		t.Errorf("expected KEY1=value1, got %s=%s", p.Vars[0].Key, p.Vars[0].Value)
	}

	if p.Vars[1].Key != "KEY2" || p.Vars[1].Value != "quoted" {
		t.Errorf("expected KEY2=quoted, got %s=%s", p.Vars[1].Key, p.Vars[1].Value)
	}

	if p.Vars[2].Key != "KEY3" || p.Vars[2].Value != "single" {
		t.Errorf("expected KEY3=single, got %s=%s", p.Vars[2].Key, p.Vars[2].Value)
	}

	if p.Vars[3].Key != "EMPTY" || p.Vars[3].Value != "" {
		t.Errorf("expected EMPTY=, got %s=%s", p.Vars[3].Key, p.Vars[3].Value)
	}
}

func TestManager_LoadNotExist(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	_, err := m.Load("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestManager_Save(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	vars := []EnvVar{
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY2", Value: "value2", Comment: "A comment"},
	}

	if err := m.Save("test", vars); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	p, err := m.Load("test")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(p.Vars) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(p.Vars))
	}

	if p.Vars[0].Key != "KEY1" || p.Vars[0].Value != "value1" {
		t.Errorf("expected KEY1=value1, got %s=%s", p.Vars[0].Key, p.Vars[0].Value)
	}
}

func TestManager_Create(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	if err := m.Create("test"); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !m.Exists("test") {
		t.Error("profile should exist after create")
	}
}

func TestManager_CreateAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	if err := m.Create("test"); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := m.Create("test"); err == nil {
		t.Error("expected error when creating existing profile")
	}
}

func TestManager_Delete(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	os.WriteFile(filepath.Join(dir, "test.env"), []byte("KEY=value"), 0644)

	if err := m.Delete("test"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if m.Exists("test") {
		t.Error("profile should not exist after delete")
	}
}

func TestManager_DeleteNotExist(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	if err := m.Delete("nonexistent"); err == nil {
		t.Error("expected error when deleting nonexistent profile")
	}
}

func TestManager_Copy(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	os.WriteFile(filepath.Join(dir, "src.env"), []byte("KEY=value"), 0644)

	if err := m.Copy("src", "dst"); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	if !m.Exists("dst") {
		t.Error("destination profile should exist")
	}

	content, _ := os.ReadFile(filepath.Join(dir, "dst.env"))
	if string(content) != "KEY=value" {
		t.Errorf("expected KEY=value, got %s", string(content))
	}
}

func TestManager_CopySrcNotExist(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	if err := m.Copy("nonexistent", "dst"); err == nil {
		t.Error("expected error when source does not exist")
	}
}

func TestManager_CopyDstExists(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	os.WriteFile(filepath.Join(dir, "src.env"), []byte("KEY=value"), 0644)
	os.WriteFile(filepath.Join(dir, "dst.env"), []byte("KEY=value"), 0644)

	if err := m.Copy("src", "dst"); err == nil {
		t.Error("expected error when destination exists")
	}
}

func TestManager_GetVars(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	content := `KEY1=value1
KEY2=value2
`
	os.WriteFile(filepath.Join(dir, "test.env"), []byte(content), 0644)

	vars, err := m.GetVars("test")
	if err != nil {
		t.Fatalf("GetVars failed: %v", err)
	}

	if vars["KEY1"] != "value1" {
		t.Errorf("expected value1, got %s", vars["KEY1"])
	}
	if vars["KEY2"] != "value2" {
		t.Errorf("expected value2, got %s", vars["KEY2"])
	}
}

func TestManager_Diff(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	contentA := `KEY1=value1
KEY2=value2
KEY3=value3
`
	contentB := `KEY1=value1
KEY2=changed
KEY4=value4
`
	os.WriteFile(filepath.Join(dir, "a.env"), []byte(contentA), 0644)
	os.WriteFile(filepath.Join(dir, "b.env"), []byte(contentB), 0644)

	added, removed, changed, err := m.Diff("a", "b")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if len(added) != 1 || added["KEY4"] != "value4" {
		t.Errorf("expected KEY4=value4 in added, got %v", added)
	}

	if len(removed) != 1 || removed["KEY3"] != "value3" {
		t.Errorf("expected KEY3=value3 in removed, got %v", removed)
	}

	if len(changed) != 1 || changed["KEY2"] != "changed" {
		t.Errorf("expected KEY2=changed in changed, got %v", changed)
	}
}

func TestManager_DiffIdentical(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	content := `KEY1=value1
`
	os.WriteFile(filepath.Join(dir, "a.env"), []byte(content), 0644)
	os.WriteFile(filepath.Join(dir, "b.env"), []byte(content), 0644)

	added, removed, changed, err := m.Diff("a", "b")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Error("expected no differences")
	}
}

func TestParseEnvFile(t *testing.T) {
	content := `# Comment
KEY1=value1
KEY2="quoted"
KEY3='single'

KEY4=value4
`

	vars := ParseEnvFile(content)

	if len(vars) != 4 {
		t.Fatalf("expected 4 vars, got %d", len(vars))
	}

	if vars[0].Key != "KEY1" || vars[0].Value != "value1" {
		t.Errorf("expected KEY1=value1, got %s=%s", vars[0].Key, vars[0].Value)
	}

	if vars[1].Key != "KEY2" || vars[1].Value != "quoted" {
		t.Errorf("expected KEY2=quoted, got %s=%s", vars[1].Key, vars[1].Value)
	}

	if vars[2].Key != "KEY3" || vars[2].Value != "single" {
		t.Errorf("expected KEY3=single, got %s=%s", vars[2].Key, vars[2].Value)
	}

	if vars[3].Key != "KEY4" || vars[3].Value != "value4" {
		t.Errorf("expected KEY4=value4, got %s=%s", vars[3].Key, vars[3].Value)
	}
}

func TestManager_ListReadDirError(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, "nonexistent"))

	_, err := m.List()
	if err == nil {
		t.Error("expected error when directory does not exist")
	}
}

func TestManager_LoadWithInvalidLine(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	content := `KEY1=value1
invalid line without equals
KEY2=value2
`
	os.WriteFile(filepath.Join(dir, "test.env"), []byte(content), 0644)

	p, err := m.Load("test")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(p.Vars) != 2 {
		t.Errorf("expected 2 vars, got %d", len(p.Vars))
	}
}

func TestManager_GetVarsError(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	_, err := m.GetVars("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestManager_DiffError(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	os.WriteFile(filepath.Join(dir, "a.env"), []byte("KEY=value"), 0644)

	_, _, _, err := m.Diff("a", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}

	_, _, _, err = m.Diff("nonexistent", "a")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestParseEnvFileEmpty(t *testing.T) {
	vars := ParseEnvFile("")
	if len(vars) != 0 {
		t.Errorf("expected 0 vars, got %d", len(vars))
	}
}

func TestParseEnvFileNoEquals(t *testing.T) {
	content := `KEY1=value1
noequals
KEY2=value2
`
	vars := ParseEnvFile(content)
	if len(vars) != 2 {
		t.Errorf("expected 2 vars, got %d", len(vars))
	}
}
