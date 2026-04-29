package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/greatbody/envman/internal/config"
	"github.com/greatbody/envman/internal/profile"
	"github.com/greatbody/envman/internal/shell"
)

type state int

const (
	stateList state = iota
	stateCreate
	stateConfirmDelete
	stateView
	stateDiff
)

type item struct {
	name    string
	loaded  bool
	default_ bool
}

func (i item) Title() string {
	marker := "  "
	if i.default_ {
		marker = "* "
	}
	suffix := ""
	if i.loaded {
		suffix = " [loaded]"
	}
	return fmt.Sprintf("%s%s%s", marker, i.name, suffix)
}

func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.name }

type Model struct {
	list         list.Model
	state        state
	names        []string
	defaultName  string
	loaded       string
	cfgMgr       *config.Manager
	profileMgr   *profile.Manager
	selected     string
	LoadOutput   string
	textInput    textinput.Model
	confirmDel   string
	err          error
	width        int
	height       int
}

func NewModel(names []string, defaultName, loaded string, cfgMgr *config.Manager, profileMgr *profile.Manager) Model {
	items := make([]list.Item, len(names))
	loadedMap := make(map[string]bool)
	if loaded != "" {
		for _, l := range splitComma(loaded) {
			loadedMap[l] = true
		}
	}

	for i, name := range names {
		items[i] = item{
			name:     name,
			loaded:   loadedMap[name],
			default_: name == defaultName,
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "envman profiles"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	ti := textinput.New()
	ti.Placeholder = "profile name"
	ti.Focus()

	return Model{
		list:        l,
		state:       stateList,
		names:       names,
		defaultName: defaultName,
		loaded:      loaded,
		cfgMgr:      cfgMgr,
		profileMgr:  profileMgr,
		textInput:   ti,
		width:       80,
		height:      24,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateList:
			return m.updateList(msg)
		case stateCreate:
			return m.updateCreate(msg)
		case stateConfirmDelete:
			return m.updateConfirmDelete(msg)
		case stateView:
			return m.updateView(msg)
		case stateDiff:
			return m.updateView(msg)
		}
	}

	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "s":
		item, ok := m.list.SelectedItem().(item)
		if ok {
			m.defaultName = item.name
			m.cfgMgr.SetDefaultProfile(item.name)
			m.refreshList()
		}
		return m, nil

	case "l":
		item, ok := m.list.SelectedItem().(item)
		if ok {
			p, err := m.profileMgr.Load(item.name)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.LoadOutput = shell.ExportWithTracking(p.Vars, item.name, m.loaded)
			return m, tea.Quit
		}
		return m, nil

	case "n":
		m.state = stateCreate
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink

	case "d":
		item, ok := m.list.SelectedItem().(item)
		if ok {
			if item.name == m.defaultName {
				m.err = fmt.Errorf("cannot delete default profile")
				return m, nil
			}
			m.state = stateConfirmDelete
			m.confirmDel = item.name
		}
		return m, nil

	case "v":
		item, ok := m.list.SelectedItem().(item)
		if ok {
			m.selected = item.name
			m.state = stateView
		}
		return m, nil

	case "c":
		item, ok := m.list.SelectedItem().(item)
		if ok {
			dst := item.name + "-copy"
			if m.profileMgr.Exists(dst) {
				dst = dst + "-1"
			}
			m.profileMgr.Copy(item.name, dst)
			m.names = append(m.names, dst)
			m.refreshList()
		}
		return m, nil

	case "D":
		m.state = stateDiff
		m.selected = ""
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := m.textInput.Value()
		if name == "" {
			m.state = stateList
			return m, nil
		}
		if m.profileMgr.Exists(name) {
			m.err = fmt.Errorf("profile %s already exists", name)
			m.state = stateList
			return m, nil
		}
		if err := m.profileMgr.Create(name); err != nil {
			m.err = err
			m.state = stateList
			return m, nil
		}
		m.names = append(m.names, name)
		m.refreshList()
		m.state = stateList
		return m, nil

	case "esc":
		m.state = stateList
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.profileMgr.Delete(m.confirmDel)
		var newNames []string
		for _, n := range m.names {
			if n != m.confirmDel {
				newNames = append(newNames, n)
			}
		}
		m.names = newNames
		m.refreshList()
		m.state = stateList
		m.confirmDel = ""
		return m, nil

	case "n", "esc":
		m.state = stateList
		m.confirmDel = ""
		return m, nil
	}

	return m, nil
}

func (m Model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = stateList
		m.selected = ""
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case stateList:
		return m.viewList()
	case stateCreate:
		return m.viewCreate()
	case stateConfirmDelete:
		return m.viewConfirmDelete()
	case stateView:
		return m.viewProfile()
	case stateDiff:
		return m.viewDiff()
	}
	return ""
}

func (m Model) viewList() string {
	var b strings.Builder
	b.WriteString(m.list.View())
	b.WriteString("\n")

	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		b.WriteString(errStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
		m.err = nil
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(helpStyle.Render("[s] set default  [l] load  [n] new  [d] delete  [v] view  [c] copy  [D] diff  [q] quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewCreate() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		titleStyle.Render("Create new profile"),
		m.textInput.View(),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("[enter] create  [esc] cancel"),
	)
}

func (m Model) viewConfirmDelete() string {
	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	return fmt.Sprintf("\n%s\n\n%s\n",
		warnStyle.Render(fmt.Sprintf("Delete profile '%s'?", m.confirmDel)),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("[y] yes  [n] no"),
	)
}

func (m Model) viewProfile() string {
	p, err := m.profileMgr.Load(m.selected)
	if err != nil {
		return fmt.Sprintf("Error loading profile: %v", err)
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n%s\n\n", titleStyle.Render(fmt.Sprintf("Profile: %s", m.selected))))

	if len(p.Vars) == 0 {
		b.WriteString("  (empty)\n")
	} else {
		for _, v := range p.Vars {
			b.WriteString(fmt.Sprintf("  %s = %s\n", v.Key, v.Value))
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("[esc] back"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewDiff() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	return fmt.Sprintf("\n%s\n\nSelect two profiles to diff (not yet implemented in TUI).\nUse: envman diff <a> <b>\n\n%s\n",
		titleStyle.Render("Diff Profiles"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("[esc] back"),
	)
}

func (m *Model) refreshList() {
	items := make([]list.Item, len(m.names))
	loadedMap := make(map[string]bool)
	if m.loaded != "" {
		for _, l := range splitComma(m.loaded) {
			loadedMap[l] = true
		}
	}

	for i, name := range m.names {
		items[i] = item{
			name:     name,
			loaded:   loadedMap[name],
			default_: name == m.defaultName,
		}
	}
	m.list.SetItems(items)
}

func splitComma(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
