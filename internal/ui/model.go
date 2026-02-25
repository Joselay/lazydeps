package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Joselay/lazydeps/internal/scanner"
)

type viewMode int

const (
	viewList viewMode = iota
	viewDetail
	viewConfirm
	viewConfirmAll
)

type tab int

const (
	tabAll tab = iota
	tabOutdated
	tabVulnerable
)

var tabNames = []string{"All", "Outdated", "Vulnerable"}

// Messages
type scanCompleteMsg struct {
	deps       []scanner.Dependency
	ecosystems []scanner.Ecosystem
}

type scanErrorMsg struct{ err error }

type updateCompleteMsg struct {
	dep scanner.Dependency
	err error
}

type updateAllCompleteMsg struct {
	updated int
	failed  int
	errors  []string
}

type Model struct {
	dir        string
	deps       []scanner.Dependency
	filtered   []scanner.Dependency
	ecosystems []scanner.Ecosystem
	cursor     int
	offset     int
	activeTab  tab
	mode       viewMode
	width      int
	height     int
	loading    bool
	spinner    spinner.Model
	filter     textinput.Model
	filtering  bool
	vulnOnly   bool
	statusMsg  string
	confirmDep *scanner.Dependency
	showHelp   bool
}

func NewModel(dir string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	ti := textinput.New()
	ti.Placeholder = "filter dependencies..."
	ti.CharLimit = 64

	return Model{
		dir:     dir,
		loading: true,
		spinner: s,
		filter:  ti,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.scanDeps(),
	)
}

func (m *Model) scanDeps() tea.Cmd {
	dir := m.dir
	return func() tea.Msg {
		deps, ecosystems, err := scanner.ScanAll(dir)
		if err != nil {
			return scanErrorMsg{err: err}
		}
		sort.Slice(deps, func(i, j int) bool {
			if deps[i].Ecosystem != deps[j].Ecosystem {
				return deps[i].Ecosystem < deps[j].Ecosystem
			}
			return deps[i].Name < deps[j].Name
		})
		return scanCompleteMsg{deps: deps, ecosystems: ecosystems}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case scanCompleteMsg:
		m.deps = msg.deps
		m.ecosystems = msg.ecosystems
		m.loading = false
		m.applyFilter()
		outdated := 0
		vuln := 0
		for _, d := range m.deps {
			if d.Outdated {
				outdated++
			}
			if d.Vulnerable {
				vuln++
			}
		}
		m.statusMsg = fmt.Sprintf("Found %d deps (%d outdated, %d vulnerable) across %d ecosystems",
			len(m.deps), outdated, vuln, len(m.ecosystems))
		return m, nil

	case scanErrorMsg:
		m.loading = false
		m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		return m, nil

	case updateCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Update failed: %v", msg.err)
		} else {
			m.statusMsg = fmt.Sprintf("Updated %s to %s", msg.dep.Name, msg.dep.Latest)
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, m.scanDeps())
		}
		return m, nil

	case updateAllCompleteMsg:
		m.loading = false
		if msg.failed > 0 {
			m.statusMsg = fmt.Sprintf("Updated %d deps, %d failed: %s",
				msg.updated, msg.failed, strings.Join(msg.errors, "; "))
		} else {
			m.statusMsg = fmt.Sprintf("Updated all %d outdated deps", msg.updated)
		}
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.scanDeps())

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.filtering {
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		m.applyFilter()
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// Handle confirmation mode (single dep)
	if m.mode == viewConfirm {
		switch msg.String() {
		case "y", "Y":
			if m.confirmDep != nil {
				dep := *m.confirmDep
				m.mode = viewList
				m.confirmDep = nil
				m.loading = true
				m.statusMsg = fmt.Sprintf("Updating %s...", dep.Name)
				dir := m.dir
				return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
					s := scanner.GetUpdater(dep.Ecosystem)
					err := s.Update(dir, dep)
					return updateCompleteMsg{dep: dep, err: err}
				})
			}
		case "n", "N", "esc":
			m.mode = viewList
			m.confirmDep = nil
		}
		return m, nil
	}

	// Handle confirmation mode (update all)
	if m.mode == viewConfirmAll {
		switch msg.String() {
		case "y", "Y":
			m.mode = viewList
			m.loading = true
			outdated := m.outdatedDeps()
			m.statusMsg = fmt.Sprintf("Updating %d deps...", len(outdated))
			dir := m.dir
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				var updated, failed int
				var errors []string
				for _, dep := range outdated {
					s := scanner.GetUpdater(dep.Ecosystem)
					if s == nil {
						failed++
						errors = append(errors, dep.Name+": no updater")
						continue
					}
					if err := s.Update(dir, dep); err != nil {
						failed++
						errors = append(errors, dep.Name+": "+err.Error())
					} else {
						updated++
					}
				}
				return updateAllCompleteMsg{updated: updated, failed: failed, errors: errors}
			})
		case "n", "N", "esc":
			m.mode = viewList
		}
		return m, nil
	}

	// Handle filter input mode
	if m.filtering {
		switch msg.String() {
		case "esc":
			m.filtering = false
			m.filter.Blur()
			m.filter.SetValue("")
			m.applyFilter()
		case "enter":
			m.filtering = false
			m.filter.Blur()
		default:
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			m.applyFilter()
			return m, cmd
		}
		return m, nil
	}

	// Handle detail view
	if m.mode == viewDetail {
		switch msg.String() {
		case "esc", "enter", "q":
			m.mode = viewList
		case "u":
			if m.cursor < len(m.filtered) && m.filtered[m.cursor].Outdated {
				dep := m.filtered[m.cursor]
				m.confirmDep = &dep
				m.mode = viewConfirm
			}
		}
		return m, nil
	}

	// List view keys
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.ensureVisible()
		}

	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
			m.ensureVisible()
		}

	case "tab", "l":
		m.activeTab = (m.activeTab + 1) % tab(len(tabNames))
		m.cursor = 0
		m.offset = 0
		m.applyFilter()

	case "shift+tab", "H":
		m.activeTab = (m.activeTab - 1 + tab(len(tabNames))) % tab(len(tabNames))
		m.cursor = 0
		m.offset = 0
		m.applyFilter()

	case "enter":
		if len(m.filtered) > 0 {
			m.mode = viewDetail
		}

	case "u":
		if m.cursor < len(m.filtered) && m.filtered[m.cursor].Outdated {
			dep := m.filtered[m.cursor]
			m.confirmDep = &dep
			m.mode = viewConfirm
		}

	case "U":
		if len(m.outdatedDeps()) > 0 {
			m.mode = viewConfirmAll
		}

	case "/":
		m.filtering = true
		m.filter.Focus()
		return m, textinput.Blink

	case "v":
		m.vulnOnly = !m.vulnOnly
		m.cursor = 0
		m.offset = 0
		m.applyFilter()

	case "r":
		m.loading = true
		m.statusMsg = "Refreshing..."
		return m, tea.Batch(m.spinner.Tick, m.scanDeps())

	case "?":
		m.showHelp = !m.showHelp
	}

	return m, nil
}

func (m *Model) applyFilter() {
	var result []scanner.Dependency

	filterText := strings.ToLower(m.filter.Value())

	for _, dep := range m.deps {
		// Tab filter
		switch m.activeTab {
		case tabOutdated:
			if !dep.Outdated {
				continue
			}
		case tabVulnerable:
			if !dep.Vulnerable {
				continue
			}
		}

		// Vuln filter
		if m.vulnOnly && !dep.Vulnerable {
			continue
		}

		// Text filter
		if filterText != "" && !strings.Contains(strings.ToLower(dep.Name), filterText) {
			continue
		}

		result = append(result, dep)
	}

	m.filtered = result

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *Model) outdatedDeps() []scanner.Dependency {
	var out []scanner.Dependency
	for _, dep := range m.deps {
		if dep.Outdated {
			out = append(out, dep)
		}
	}
	return out
}

func (m *Model) ensureVisible() {
	tableHeight := m.tableHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+tableHeight {
		m.offset = m.cursor - tableHeight + 1
	}
}

func (m *Model) tableHeight() int {
	// Reserve lines for: title(1) + tabs(1) + header(2) + footer(2) + status(1) + padding(2)
	h := m.height - 9
	if h < 5 {
		h = 5
	}
	return h
}
