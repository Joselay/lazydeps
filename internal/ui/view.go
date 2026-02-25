package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Joselay/lazydeps/internal/scanner"
)

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var sections []string

	// Title bar
	sections = append(sections, m.renderTitle())

	if m.loading {
		spinner := m.spinner.View()
		msg := lipgloss.NewStyle().Foreground(colorMuted).Render("Scanning dependencies...")
		sections = append(sections, fmt.Sprintf("\n  %s %s\n", spinner, msg))
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

	// Ecosystem badges
	var badges []string
	for _, eco := range m.ecosystems {
		badges = append(badges, m.titleBadge(eco))
	}

	ecoStr := ""
	if len(badges) > 0 {
		ecoStr = " " + strings.Join(badges, " ")
	}

	right := subtitleStyle.Render(m.dir)
	gap := m.width - lipgloss.Width(title) - lipgloss.Width(ecoStr) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	return title + ecoStr + strings.Repeat(" ", gap) + right
}

func (m Model) titleBadge(eco scanner.Ecosystem) string {
	switch eco {
	case scanner.EcosystemGo:
		return lipgloss.NewStyle().Foreground(colorGo).Bold(true).Render("Go")
	case scanner.EcosystemNpm:
		return lipgloss.NewStyle().Foreground(colorNpm).Bold(true).Render("npm")
	case scanner.EcosystemPip:
		return lipgloss.NewStyle().Foreground(colorPip).Bold(true).Render("pip")
	case scanner.EcosystemCargo:
		return lipgloss.NewStyle().Foreground(colorCargo).Bold(true).Render("cargo")
	case scanner.EcosystemBun:
		return lipgloss.NewStyle().Foreground(colorBun).Bold(true).Render("bun")
	default:
		return "?"
	}
}

func (m Model) renderTabs() string {
	var tabs []string
	for i, name := range tabNames {
		count := m.countForTab(tab(i))
		label := fmt.Sprintf("%s %d", name, count)
		if tab(i) == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(label))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
	// Extend the bottom border across the full width
	rowWidth := lipgloss.Width(row)
	if rowWidth < m.width {
		fill := tabBarStyle.Width(m.width - rowWidth).Render("")
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, fill)
	}
	return row
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
		empty := lipgloss.NewStyle().Foreground(colorMuted).Italic(true).Render("No dependencies found for this view.")
		return "\n  " + empty + "\n"
	}

	var b strings.Builder

	// Header
	header := fmt.Sprintf("  %-5s %-40s %-16s %-16s %s",
		"", "PACKAGE", "CURRENT", "LATEST", "STATUS")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	tableHeight := m.tableHeight()
	end := m.offset + tableHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := m.offset; i < end; i++ {
		dep := m.filtered[i]
		isAlt := (i-m.offset)%2 == 1
		row := m.renderRow(dep, i == m.cursor, isAlt)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.filtered) > tableHeight {
		pct := 0
		if len(m.filtered)-tableHeight > 0 {
			pct = m.offset * 100 / (len(m.filtered) - tableHeight)
		}
		scrollInfo := fmt.Sprintf("%d–%d of %d", m.offset+1, end, len(m.filtered))

		// Mini scrollbar
		barHeight := tableHeight
		thumbPos := 0
		if len(m.filtered)-tableHeight > 0 {
			thumbPos = m.offset * (barHeight - 1) / (len(m.filtered) - tableHeight)
		}
		_ = thumbPos // scrollbar rendered inline in rows if desired

		right := scrollStyle.Render(scrollInfo)
		bar := scrollStyle.Render(fmt.Sprintf(" %d%%", pct))
		b.WriteString(strings.Repeat(" ", max(2, m.width-lipgloss.Width(right)-lipgloss.Width(bar)-2)))
		b.WriteString(right)
		b.WriteString(bar)
	}

	return b.String()
}

func (m Model) renderRow(dep scanner.Dependency, selected bool, isAlt bool) string {
	name := dep.Name
	if len(name) > 38 {
		name = name[:35] + "..."
	}

	current := dep.Current
	if len(current) > 14 {
		current = current[:14]
	}
	latest := dep.Latest
	if len(latest) > 14 {
		latest = latest[:14]
	}

	ecoLabel := m.ecosystemLabel(dep.Ecosystem)

	// Status with icons
	var statusLabel string
	if dep.Vulnerable {
		statusLabel = "⚠ VULN"
	} else if dep.Outdated {
		statusLabel = "⬆ UPDATE"
	} else {
		statusLabel = "✓ OK"
	}

	if selected {
		prefix := "▸"
		row := fmt.Sprintf("%s %-5s %-40s %-16s %-16s %s",
			prefix, ecoLabel, name, current, latest, statusLabel)
		return selectedRowStyle.Width(min(m.width, 110)).Render(row)
	}

	// Unselected: colored ecosystem + status
	eco := m.ecosystemBadge(dep.Ecosystem)
	var status string
	if dep.Vulnerable {
		status = vulnerableStyle.Render("⚠ VULN")
	} else if dep.Outdated {
		status = outdatedStyle.Render("⬆ UPDATE")
	} else {
		status = upToDateStyle.Render("✓ OK")
	}

	row := fmt.Sprintf("  %-5s %-40s %-16s %-16s %s",
		eco, name, current, latest, status)

	if isAlt {
		return altRowStyle.Width(min(m.width, 110)).Render(row)
	}
	return normalRowStyle.Render(row)
}

func (m Model) ecosystemLabel(eco scanner.Ecosystem) string {
	switch eco {
	case scanner.EcosystemGo:
		return "Go"
	case scanner.EcosystemNpm:
		return "npm"
	case scanner.EcosystemPip:
		return "pip"
	case scanner.EcosystemCargo:
		return "cargo"
	case scanner.EcosystemBun:
		return "bun"
	default:
		return "?"
	}
}

func (m Model) ecosystemBadge(eco scanner.Ecosystem) string {
	switch eco {
	case scanner.EcosystemGo:
		return lipgloss.NewStyle().Foreground(colorGo).Render("Go")
	case scanner.EcosystemNpm:
		return lipgloss.NewStyle().Foreground(colorNpm).Render("npm")
	case scanner.EcosystemPip:
		return lipgloss.NewStyle().Foreground(colorPip).Render("pip")
	case scanner.EcosystemCargo:
		return lipgloss.NewStyle().Foreground(colorCargo).Render("cargo")
	case scanner.EcosystemBun:
		return lipgloss.NewStyle().Foreground(colorBun).Render("bun")
	default:
		return "?"
	}
}

func (m Model) renderDetail() string {
	if m.cursor >= len(m.filtered) {
		return ""
	}
	dep := m.filtered[m.cursor]

	var lines []string

	// Package name
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(colorWhite)
	lines = append(lines, nameStyle.Render(dep.Name))
	lines = append(lines, "")

	// Info rows
	lines = append(lines, detailLabelStyle.Render("Ecosystem")+detailValueStyle.Render(string(dep.Ecosystem)))
	lines = append(lines, detailLabelStyle.Render("Current")+detailValueStyle.Render(dep.Current))
	lines = append(lines, detailLabelStyle.Render("Latest")+detailValueStyle.Render(dep.Latest))

	depType := "direct"
	if dep.Indirect {
		depType = "dev / indirect"
	}
	lines = append(lines, detailLabelStyle.Render("Type")+detailValueStyle.Render(depType))

	if dep.Outdated {
		lines = append(lines, "")
		arrow := outdatedStyle.Render(fmt.Sprintf("⬆ Update available: %s → %s", dep.Current, dep.Latest))
		lines = append(lines, arrow)
	}

	if dep.Vulnerable {
		lines = append(lines, "")
		lines = append(lines, vulnerableStyle.Render("⚠ VULNERABILITY DETECTED"))
		if dep.VulnInfo != "" {
			lines = append(lines, lipgloss.NewStyle().Foreground(colorDanger).Render("  "+dep.VulnInfo))
		}
	}

	lines = append(lines, "")
	hint := helpStyle.Render("u update  •  esc back")
	lines = append(lines, hint)

	content := strings.Join(lines, "\n")
	panel := detailBorderStyle.Width(min(m.width-4, 64)).Render(content)
	return "\n" + panel + "\n"
}

func (m Model) renderConfirm() string {
	if m.confirmDep == nil {
		return ""
	}
	prompt := fmt.Sprintf("  Update %s from %s → %s? [y/N]",
		m.confirmDep.Name, m.confirmDep.Current, m.confirmDep.Latest)
	return "\n" + promptStyle.Render(prompt)
}

func (m Model) renderConfirmAll() string {
	outdated := m.outdatedDeps()
	prompt := fmt.Sprintf("  Update all %d outdated dependencies? [y/N]", len(outdated))
	return "\n" + promptStyle.Render(prompt)
}

func (m Model) renderStatusBar() string {
	if m.statusMsg == "" {
		return ""
	}
	return statusBarStyle.Width(m.width).Render(m.statusMsg)
}

func (m Model) renderHelp() string {
	if m.showHelp {
		bindings := []struct{ key, desc string }{
			{"↑/k  ↓/j", "Navigate list"},
			{"tab / S-tab", "Switch tabs"},
			{"enter", "View details"},
			{"u", "Update dependency"},
			{"U", "Update all outdated"},
			{"/", "Filter by name"},
			{"v", "Toggle vuln only"},
			{"r", "Refresh scan"},
			{"?", "Toggle help"},
			{"q", "Quit"},
		}

		var rows []string
		for _, b := range bindings {
			rows = append(rows, helpKeyStyle.Render(b.key)+helpDescStyle.Render(b.desc))
		}

		content := strings.Join(rows, "\n")
		panel := helpPanelStyle.Width(min(m.width-4, 40)).Render(content)
		return "\n" + panel
	}

	keys := []string{"↑↓", "tab", "enter", "u", "U", "/", "?", "q"}
	labels := []string{"navigate", "tabs", "details", "update", "update all", "filter", "help", "quit"}

	var parts []string
	for i := range keys {
		k := lipgloss.NewStyle().Foreground(colorTextBright).Render(keys[i])
		parts = append(parts, k+" "+helpDescStyle.Render(labels[i]))
	}
	return helpStyle.Render("  " + strings.Join(parts, "  "))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
