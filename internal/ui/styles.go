package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7C3AED")
	colorSecondary = lipgloss.Color("#06B6D4")
	colorSuccess   = lipgloss.Color("#10B981")
	colorWarning   = lipgloss.Color("#F59E0B")
	colorDanger    = lipgloss.Color("#EF4444")
	colorMuted     = lipgloss.Color("#6B7280")
	colorText      = lipgloss.Color("#E5E7EB")
	colorBg        = lipgloss.Color("#1F2937")
	colorBgAlt     = lipgloss.Color("#374151")
	colorWhite     = lipgloss.Color("#FFFFFF")

	// App chrome
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorPrimary).
			Padding(0, 2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	// Tabs
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorPrimary).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Padding(0, 2)

	// Table rows
	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite).
				Background(colorPrimary)

	normalRowStyle = lipgloss.NewStyle().
			Foreground(colorText)

	// Status indicators
	outdatedStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	vulnerableStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	upToDateStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Ecosystem badges
	goBadge    = lipgloss.NewStyle().Foreground(colorSecondary).Bold(true).Render("GO")
	npmBadge   = lipgloss.NewStyle().Foreground(lipgloss.Color("#CB3837")).Bold(true).Render("NPM")
	pipBadge   = lipgloss.NewStyle().Foreground(lipgloss.Color("#3776AB")).Bold(true).Render("PIP")
	cargoBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("#DEA584")).Bold(true).Render("RST")
	bunBadge   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBF0DF")).Bold(true).Render("BUN")

	// Footer
	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorBgAlt).
			Padding(0, 1)

	// Detail panel
	detailBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(1, 2)

	// Spinner
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)

	// Confirmation prompt
	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWarning).
			Padding(0, 1)
)
