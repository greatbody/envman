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

const (
	maxWidth = 50
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
	name     string
	loaded   bool
	default_ bool
}

func (i item) Title() string {
	marker := "  "
	if i.default_ {
		marker = "● "
	}
	suffix := ""
	if i.loaded {
		suffix = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("(loaded)")
	}
	return marker + i.name + suffix
}

func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.name }

type Model struct {
	list       list.Model
	state      state
	names      []string
	defaultName string
	loaded     string
	cfgMgr     *config.Manager
	profileMgr *profile.Manager
	selected   string
	LoadOutput string
	textInput  textinput.Model
	confirmDel string
	err        error
	width      int
	height     int
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

	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(lipgloss.Color("250"))
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(lipgloss.Color("240"))
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("170")).BorderLeft(true).BorderForeground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("170"))

	l := list.New(items, delegate, maxWidth, 0)
	l.Title = "envman"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowTitle(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		PaddingRight(1)

	ti := textinput.New()
	ti.Placeholder = "profile name"
	ti.Focus()

	return Model{
		list:       l,
		state:      stateList,
		names:      names,
		defaultName: defaultName,
		loaded:     loaded,
		cfgMgr:     cfgMgr,
		profileMgr: profileMgr,
		textInput:  ti,
		width:      maxWidth,
		height:     20,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth)
		m.height = msg.Height
		m.list.SetWidth(m.width)
		m.list.SetHeight(min(msg.Height-4, len(m.names)+2))
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateList:
			return m.updateList(msg)
		case stateCreate:
			return m.updateCreate(msg)
		case stateConfirmDelete:
			return m.updateConfirmDelete(msg)
		case stateView, stateDiff:
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
		selected, ok := m.list.SelectedItem().(item)
		if ok {
			m.defaultName = selected.name
			m.cfgMgr.SetDefaultProfile(selected.name)
			m.refreshList()
		}
		return m, nil

	case "l":
		selected, ok := m.list.SelectedItem().(item)
		if ok {
			p, err := m.profileMgr.Load(selected.name)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.LoadOutput = shell.ExportWithTracking(p.Vars, selected.name, m.loaded)
			return m, tea.Quit
		}
		return m, nil

	case "n":
		m.state = stateCreate
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink

	case "d":
		selected, ok := m.list.SelectedItem().(item)
		if ok {
			if selected.name == m.defaultName {
				m.err = fmt.Errorf("cannot delete default profile")
				return m, nil
			}
			m.state = stateConfirmDelete
			m.confirmDel = selected.name
		}
		return m, nil

	case "v":
		selected, ok := m.list.SelectedItem().(item)
		if ok {
			m.selected = selected.name
			m.state = stateView
		}
		return m, nil

	case "c":
		selected, ok := m.list.SelectedItem().(item)
		if ok {
			dst := selected.name + "-copy"
			if m.profileMgr.Exists(dst) {
				dst = dst + "-1"
			}
			m.profileMgr.Copy(selected.name, dst)
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

	if m.err != nil {
		b.WriteString("\n")
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		b.WriteString(errStyle.Render(fmt.Sprintf("  %v", m.err)))
		m.err = nil
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(1)
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("s:set  l:load  n:new  d:del  v:view  c:copy  q:quit"))

	return b.String()
}

func (m Model) viewCreate() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		PaddingLeft(1)

	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(1)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Create profile"))
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")
	b.WriteString(promptStyle.Render("enter:create  esc:cancel"))

	return b.String()
}

func (m Model) viewConfirmDelete() string {
	warnStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("1")).
		PaddingLeft(1)

	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(1)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(warnStyle.Render(fmt.Sprintf("Delete '%s'?", m.confirmDel)))
	b.WriteString("\n\n")
	b.WriteString(promptStyle.Render("y:yes  n:no"))

	return b.String()
}

func (m Model) viewProfile() string {
	p, err := m.profileMgr.Load(m.selected)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		PaddingLeft(1)

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("178")).PaddingLeft(3)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).PaddingLeft(0)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render(m.selected))
	b.WriteString("\n\n")

	if len(p.Vars) == 0 {
		b.WriteString("  (empty)")
	} else {
		for _, v := range p.Vars {
			b.WriteString(keyStyle.Render(v.Key))
			b.WriteString(valStyle.Render(" = " + v.Value))
			b.WriteString("\n")
		}
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(1)
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("esc:back"))

	return b.String()
}

func (m Model) viewDiff() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		PaddingLeft(1)

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(1)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Diff"))
	b.WriteString("\n\n")
	b.WriteString("  Use: envman diff <a> <b>")
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("esc:back"))

	return b.String()
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
