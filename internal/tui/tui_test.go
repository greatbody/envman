package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/greatbody/envman/internal/config"
	"github.com/greatbody/envman/internal/profile"
)

func TestItem_Title(t *testing.T) {
	tests := []struct {
		name     string
		item     item
		expected string
	}{
		{
			name:     "normal",
			item:     item{name: "test"},
			expected: "  test",
		},
		{
			name:     "default",
			item:     item{name: "test", default_: true},
			expected: "* test",
		},
		{
			name:     "loaded",
			item:     item{name: "test", loaded: true},
			expected: "  test [loaded]",
		},
		{
			name:     "default and loaded",
			item:     item{name: "test", default_: true, loaded: true},
			expected: "* test [loaded]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.Title()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestItem_Description(t *testing.T) {
	i := item{name: "test"}
	if i.Description() != "" {
		t.Error("expected empty description")
	}
}

func TestItem_FilterValue(t *testing.T) {
	i := item{name: "test"}
	if i.FilterValue() != "test" {
		t.Errorf("expected test, got %s", i.FilterValue())
	}
}

func TestNewModel(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	names := []string{"default", "work"}
	model := NewModel(names, "default", "work", cfgMgr, profileMgr)

	if model.defaultName != "default" {
		t.Errorf("expected default, got %s", model.defaultName)
	}

	if model.loaded != "work" {
		t.Errorf("expected work, got %s", model.loaded)
	}

	if len(model.names) != 2 {
		t.Errorf("expected 2 names, got %d", len(model.names))
	}
}

func TestSplitComma(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{",,,", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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

func TestModelInit(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	cmd := model.Init()
	if cmd != nil {
		t.Error("expected nil cmd from Init")
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updated, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	m := updated.(Model)
	if m.width != 100 || m.height != 30 {
		t.Errorf("expected 100x30, got %dx%d", m.width, m.height)
	}
}

func TestModelUpdateQuit(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updated, cmd := model.Update(msg)
	if cmd == nil {
		t.Error("expected quit cmd")
	}
	_ = updated
}

func TestModelUpdateQuitCtrlC(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := model.Update(msg)
	if cmd == nil {
		t.Error("expected quit cmd")
	}
}

func TestModelUpdateSetDefault(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())
	profileMgr.Create("work")

	model := NewModel([]string{"default", "work"}, "default", "", cfgMgr, profileMgr)

	// Simulate selecting "work" by setting the list index
	model.list.Select(1)

	// Press 's' to set default
	sMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	model.Update(sMsg)

	defaultName, _ := cfgMgr.GetDefaultProfile()
	if defaultName != "work" {
		t.Errorf("expected work, got %s", defaultName)
	}
}

func TestModelUpdateLoad(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())
	profileMgr.Save("work", []profile.EnvVar{{Key: "KEY1", Value: "value1"}})

	model := NewModel([]string{"default", "work"}, "default", "", cfgMgr, profileMgr)

	// Select "work"
	model.list.Select(1)

	// Press 'l' to load
	lMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updated, cmd := model.Update(lMsg)
	m := updated.(Model)

	if cmd == nil {
		t.Error("expected quit cmd")
	}
	if m.LoadOutput == "" {
		t.Error("expected load output")
	}
}

func TestModelUpdateCreate(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)

	// Press 'n' to create
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updated, _ := model.Update(nMsg)
	m := updated.(Model)

	if m.state != stateCreate {
		t.Errorf("expected stateCreate, got %d", m.state)
	}

	// Type name
	for _, c := range "newprofile" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{c}}
		updated, _ = m.Update(msg)
		m = updated.(Model)
	}

	// Press enter
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ = m.Update(enterMsg)
	m = updated.(Model)

	if m.state != stateList {
		t.Errorf("expected stateList, got %d", m.state)
	}

	if !profileMgr.Exists("newprofile") {
		t.Error("expected newprofile to exist")
	}
}

func TestModelUpdateCreateExists(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())
	profileMgr.Create("existing")

	model := NewModel([]string{"default", "existing"}, "default", "", cfgMgr, profileMgr)

	// Press 'n'
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updated, _ := model.Update(nMsg)
	m := updated.(Model)

	// Type "existing"
	for _, c := range "existing" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{c}}
		updated, _ = m.Update(msg)
		m = updated.(Model)
	}

	// Press enter
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ = m.Update(enterMsg)
	m = updated.(Model)

	if m.err == nil {
		t.Error("expected error for existing profile")
	}
}

func TestModelUpdateCreateEmpty(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)

	// Press 'n'
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updated, _ := model.Update(nMsg)
	m := updated.(Model)

	// Press enter with empty name
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ = m.Update(enterMsg)
	m = updated.(Model)

	if m.state != stateList {
		t.Errorf("expected stateList, got %d", m.state)
	}
}

func TestModelUpdateCreateCancel(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)

	// Press 'n'
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updated, _ := model.Update(nMsg)
	m := updated.(Model)

	// Press esc
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = m.Update(escMsg)
	m = updated.(Model)

	if m.state != stateList {
		t.Errorf("expected stateList, got %d", m.state)
	}
}

func TestModelUpdateDelete(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())
	profileMgr.Create("todelete")

	model := NewModel([]string{"default", "todelete"}, "default", "", cfgMgr, profileMgr)

	// Select "todelete"
	model.list.Select(1)

	// Press 'd'
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	updated, _ := model.Update(dMsg)
	m := updated.(Model)

	if m.state != stateConfirmDelete {
		t.Errorf("expected stateConfirmDelete, got %d", m.state)
	}

	// Press 'y' to confirm
	yMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	updated, _ = m.Update(yMsg)
	m = updated.(Model)

	if m.state != stateList {
		t.Errorf("expected stateList, got %d", m.state)
	}

	if profileMgr.Exists("todelete") {
		t.Error("expected todelete to be deleted")
	}
}

func TestModelUpdateDeleteDefault(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)

	// Press 'd' on default
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	updated, _ := model.Update(dMsg)
	m := updated.(Model)

	if m.err == nil {
		t.Error("expected error when deleting default")
	}
}

func TestModelUpdateDeleteCancel(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())
	profileMgr.Create("todelete")

	model := NewModel([]string{"default", "todelete"}, "default", "", cfgMgr, profileMgr)

	// Select "todelete"
	model.list.Select(1)

	// Press 'd'
	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	updated, _ := model.Update(dMsg)
	m := updated.(Model)

	// Press 'n' to cancel
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updated, _ = m.Update(nMsg)
	m = updated.(Model)

	if m.state != stateList {
		t.Errorf("expected stateList, got %d", m.state)
	}

	if !profileMgr.Exists("todelete") {
		t.Error("expected todelete to still exist")
	}
}

func TestModelUpdateDeleteEsc(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())
	profileMgr.Create("todelete")

	model := NewModel([]string{"default", "todelete"}, "default", "", cfgMgr, profileMgr)

	model.list.Select(1)

	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	updated, _ := model.Update(dMsg)
	m := updated.(Model)

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = m.Update(escMsg)
	m = updated.(Model)

	if m.state != stateList {
		t.Errorf("expected stateList, got %d", m.state)
	}
}

func TestModelUpdateView(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())
	profileMgr.Save("test", []profile.EnvVar{{Key: "KEY1", Value: "value1"}})

	model := NewModel([]string{"default", "test"}, "default", "", cfgMgr, profileMgr)

	model.list.Select(1)

	vMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}
	updated, _ := model.Update(vMsg)
	m := updated.(Model)

	if m.state != stateView {
		t.Errorf("expected stateView, got %d", m.state)
	}

	if m.selected != "test" {
		t.Errorf("expected test, got %s", m.selected)
	}
}

func TestModelUpdateViewBack(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	model.state = stateView
	model.selected = "default"

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := model.Update(escMsg)
	m := updated.(Model)

	if m.state != stateList {
		t.Errorf("expected stateList, got %d", m.state)
	}
}

func TestModelUpdateViewQ(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	model.state = stateView
	model.selected = "default"

	qMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updated, _ := model.Update(qMsg)
	m := updated.(Model)

	if m.state != stateList {
		t.Errorf("expected stateList, got %d", m.state)
	}
}

func TestModelUpdateCopy(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())
	profileMgr.Save("src", []profile.EnvVar{{Key: "KEY1", Value: "value1"}})

	model := NewModel([]string{"default", "src"}, "default", "", cfgMgr, profileMgr)

	model.list.Select(1)

	cMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	updated, _ := model.Update(cMsg)
	m := updated.(Model)

	if !profileMgr.Exists("src-copy") {
		t.Error("expected src-copy to exist")
	}
	_ = m
}

func TestModelUpdateDiff(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)

	dMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}}
	updated, _ := model.Update(dMsg)
	m := updated.(Model)

	if m.state != stateDiff {
		t.Errorf("expected stateDiff, got %d", m.state)
	}
}

func TestModelUpdateDiffBack(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	model.state = stateDiff

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := model.Update(escMsg)
	m := updated.(Model)

	if m.state != stateList {
		t.Errorf("expected stateList, got %d", m.state)
	}
}

func TestModelUpdateOtherKeyInList(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)

	// Press an unhandled key (like 'x')
	xMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ := model.Update(xMsg)
	m := updated.(Model)

	if m.state != stateList {
		t.Errorf("expected stateList, got %d", m.state)
	}
}

func TestModelView(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)

	view := model.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestModelViewCreate(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	model.state = stateCreate

	view := model.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestModelViewConfirmDelete(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	model.state = stateConfirmDelete
	model.confirmDel = "test"

	view := model.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestModelViewProfile(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())
	profileMgr.Save("test", []profile.EnvVar{{Key: "KEY1", Value: "value1"}})

	model := NewModel([]string{"default", "test"}, "default", "", cfgMgr, profileMgr)
	model.state = stateView
	model.selected = "test"

	view := model.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestModelViewProfileEmpty(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())
	profileMgr.Create("empty")

	model := NewModel([]string{"default", "empty"}, "default", "", cfgMgr, profileMgr)
	model.state = stateView
	model.selected = "empty"

	view := model.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestModelViewProfileError(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	model.state = stateView
	model.selected = "nonexistent"

	view := model.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestModelViewDiff(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	model.state = stateDiff

	view := model.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestModelViewUnknownState(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	model.state = state(99)

	view := model.View()
	if view != "" {
		t.Errorf("expected empty view, got %q", view)
	}
}

func TestModelUpdateCreateNonListState(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManagerWithBaseDir(dir)
	cfgMgr.Init()
	profileMgr := profile.NewManager(cfgMgr.ProfilesDir())

	model := NewModel([]string{"default"}, "default", "", cfgMgr, profileMgr)
	model.state = state(99)

	xMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ := model.Update(xMsg)
	m := updated.(Model)

	if m.state != state(99) {
		t.Errorf("expected state 99, got %d", m.state)
	}
}
