package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/smaetongmenglay/lazydeps/internal/scanner"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var sections []string

	// Title bar
	sections = append(sections, m.renderTitle())

	if m.loading {
		sections = append(sections, fmt.Sprintf("\n  %s Scanning dependencies...\n", m.spinner.View()))
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	// Tabs
	sections = append(sections, m.renderTabs())

	// Filter bar (if active)
	if m.filtering || m.filter.Value() != "" {
		sections = append(sections, m.renderFilterBar())
	}

	// Main content
	switch m.mode {
	case viewDetail:
		sections = append(sections, m.renderDetail())
	case viewConfirm:
		sections = append(sections, m.renderTable())
		sections = append(sections, m.renderConfirm())
	case viewConfirmAll:
		sections = append(sections, m.renderTable())
		sections = append(sections, m.renderConfirmAll())
	default:
		sections = append(sections, m.renderTable())
	}

	// Status bar
	sections = append(sections, m.renderStatusBar())

	// Help
	sections = append(sections, m.renderHelp())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderTitle() string {
	title := titleStyle.Render(" lazydeps ")
	ecosystems := []string{}
	for _, eco := range m.ecosystems {
		switch eco {
		case scanner.EcosystemGo:
			ecosystems = append(ecosystems, goBadge)
		case scanner.EcosystemNpm:
			ecosystems = append(ecosystems, npmBadge)
		case scanner.EcosystemPip:
			ecosystems = append(ecosystems, pipBadge)
		case scanner.EcosystemCargo:
			ecosystems = append(ecosystems, cargoBadge)
		case scanner.EcosystemBun:
			ecosystems = append(ecosystems, bunBadge)
		}
	}

	ecoStr := ""
	if len(ecosystems) > 0 {
		ecoStr = " " + strings.Join(ecosystems, " ")
	}

	right := subtitleStyle.Render(m.dir)
	gap := m.width - lipgloss.Width(title) - lipgloss.Width(ecoStr) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	return title + ecoStr + strings.Repeat(" ", gap) + right
}

func (m Model) renderTabs() string {
	var tabs []string
	for i, name := range tabNames {
		count := m.countForTab(tab(i))
		label := fmt.Sprintf("%s (%d)", name, count)
		if tab(i) == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(label))
		}
	}
	return " " + strings.Join(tabs, " ")
}

func (m Model) countForTab(t tab) int {
	count := 0
	for _, dep := range m.deps {
		switch t {
		case tabAll:
			count++
		case tabOutdated:
			if dep.Outdated {
				count++
			}
		case tabVulnerable:
			if dep.Vulnerable {
				count++
			}
		}
	}
	return count
}

func (m Model) renderFilterBar() string {
	return "  " + m.filter.View()
}

func (m Model) renderTable() string {
	if len(m.filtered) == 0 {
		return "\n  " + subtitleStyle.Render("No dependencies found for this view.") + "\n"
	}

	var b strings.Builder

	// Header
	header := fmt.Sprintf("  %-4s %-40s %-16s %-16s %s",
		"ECO", "PACKAGE", "CURRENT", "LATEST", "STATUS")
	b.WriteString(lipgloss.NewStyle().
		Foreground(colorMuted).
		Bold(true).
		Render(header))
	b.WriteString("\n")
	b.WriteString("  " + strings.Repeat("─", min(m.width-4, 100)))
	b.WriteString("\n")

	tableHeight := m.tableHeight()
	end := m.offset + tableHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := m.offset; i < end; i++ {
		dep := m.filtered[i]
		row := m.renderRow(dep, i == m.cursor)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.filtered) > tableHeight {
		pct := 0
		if len(m.filtered)-tableHeight > 0 {
			pct = m.offset * 100 / (len(m.filtered) - tableHeight)
		}
		scrollInfo := fmt.Sprintf("  %d-%d of %d (%d%%)",
			m.offset+1, end, len(m.filtered), pct)
		b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render(scrollInfo))
	}

	return b.String()
}

func (m Model) renderRow(dep scanner.Dependency, selected bool) string {
	name := dep.Name
	if len(name) > 38 {
		name = name[:35] + "..."
	}

	current := dep.Current
	latest := dep.Latest

	ecoLabel := m.ecosystemLabel(dep.Ecosystem)
	statusLabel := "OK"
	if dep.Vulnerable {
		statusLabel = "VULN"
	} else if dep.Outdated {
		statusLabel = "UPDATE"
	}

	if selected {
		// Plain text for selected row — no nested styles so background covers full width
		row := fmt.Sprintf("▸ %-4s %-40s %-16s %-16s %s",
			ecoLabel, name, current, latest, statusLabel)
		return selectedRowStyle.Width(min(m.width, 104)).Render(row)
	}

	// Unselected: use colored badges and status
	eco := m.ecosystemBadge(dep.Ecosystem)
	status := ""
	if dep.Vulnerable {
		status = vulnerableStyle.Render(statusLabel)
	} else if dep.Outdated {
		status = outdatedStyle.Render(statusLabel)
	} else {
		status = upToDateStyle.Render(statusLabel)
	}

	row := fmt.Sprintf("  %-4s %-40s %-16s %-16s %s",
		eco, name, current, latest, status)
	return normalRowStyle.Render(row)
}

func (m Model) ecosystemLabel(eco scanner.Ecosystem) string {
	switch eco {
	case scanner.EcosystemGo:
		return "GO"
	case scanner.EcosystemNpm:
		return "NPM"
	case scanner.EcosystemPip:
		return "PIP"
	case scanner.EcosystemCargo:
		return "RST"
	case scanner.EcosystemBun:
		return "BUN"
	default:
		return "???"
	}
}

func (m Model) ecosystemBadge(eco scanner.Ecosystem) string {
	switch eco {
	case scanner.EcosystemGo:
		return lipgloss.NewStyle().Foreground(colorSecondary).Render("GO")
	case scanner.EcosystemNpm:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#CB3837")).Render("NPM")
	case scanner.EcosystemPip:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3776AB")).Render("PIP")
	case scanner.EcosystemCargo:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#DEA584")).Render("RST")
	case scanner.EcosystemBun:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FBF0DF")).Render("BUN")
	default:
		return "???"
	}
}

func (m Model) renderDetail() string {
	if m.cursor >= len(m.filtered) {
		return ""
	}
	dep := m.filtered[m.cursor]

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Render(dep.Name))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Ecosystem:  %s", string(dep.Ecosystem)))
	lines = append(lines, fmt.Sprintf("Current:    %s", dep.Current))
	lines = append(lines, fmt.Sprintf("Latest:     %s", dep.Latest))

	if dep.Indirect {
		lines = append(lines, fmt.Sprintf("Type:       %s", "indirect/dev"))
	} else {
		lines = append(lines, fmt.Sprintf("Type:       %s", "direct"))
	}

	if dep.Outdated {
		lines = append(lines, "")
		lines = append(lines, outdatedStyle.Render(fmt.Sprintf("⚠ Update available: %s → %s", dep.Current, dep.Latest)))
	}

	if dep.Vulnerable {
		lines = append(lines, "")
		lines = append(lines, vulnerableStyle.Render("✗ VULNERABILITY DETECTED"))
		if dep.VulnInfo != "" {
			lines = append(lines, vulnerableStyle.Render("  "+dep.VulnInfo))
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("Press u to update • esc to go back"))

	content := strings.Join(lines, "\n")
	panel := detailBorderStyle.Width(min(m.width-4, 70)).Render(content)
	return "\n" + panel + "\n"
}

func (m Model) renderConfirm() string {
	if m.confirmDep == nil {
		return ""
	}
	prompt := fmt.Sprintf("  Update %s from %s to %s? [y/N]",
		m.confirmDep.Name, m.confirmDep.Current, m.confirmDep.Latest)
	return "\n" + promptStyle.Render(prompt)
}

func (m Model) renderConfirmAll() string {
	outdated := m.outdatedDeps()
	prompt := fmt.Sprintf("  Update all %d outdated deps? [y/N]", len(outdated))
	return "\n" + promptStyle.Render(prompt)
}

func (m Model) renderStatusBar() string {
	if m.statusMsg == "" {
		return ""
	}
	bar := statusBarStyle.Width(m.width).Render(m.statusMsg)
	return bar
}

func (m Model) renderHelp() string {
	if m.showHelp {
		help := []string{
			"",
			helpStyle.Render("  Keybindings:"),
			helpStyle.Render("  ↑/k, ↓/j    Navigate        tab/l, S-tab/H  Switch tabs"),
			helpStyle.Render("  enter        View details    u               Update dep"),
		helpStyle.Render("  U            Update all      v               Vuln only"),
			helpStyle.Render("  /            Filter"),
			helpStyle.Render("  r            Refresh         ?               Toggle help"),
			helpStyle.Render("  q, ctrl+c    Quit"),
			"",
		}
		return strings.Join(help, "\n")
	}
	return helpStyle.Render("  ↑↓ navigate • tab switch • enter details • u update • U update all • / filter • ? help • q quit")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
