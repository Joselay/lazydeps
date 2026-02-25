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
	colorText      = lipgloss.Color("#D1D5DB")
	colorTextBright = lipgloss.Color("#F3F4F6")
	colorBgDark    = lipgloss.Color("#111827")
	colorBgAlt     = lipgloss.Color("#1F2937")
	colorBgRow     = lipgloss.Color("#1a1f2e")
	colorWhite     = lipgloss.Color("#FFFFFF")
	colorDim       = lipgloss.Color("#4B5563")

	// Title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorPrimary).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Tabs
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1).
			BorderBottom(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderBottomForeground(colorPrimary)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Padding(0, 1).
				BorderBottom(true).
				BorderStyle(lipgloss.HiddenBorder())

	tabBarStyle = lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottomForeground(colorDim)

	// Table
	headerStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Bold(true)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite).
				Background(colorPrimary)

	normalRowStyle = lipgloss.NewStyle().
			Foreground(colorText)

	altRowStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorBgRow)

	// Status indicators
	outdatedStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	vulnerableStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	upToDateStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Ecosystem colors for badges
	colorGo    = lipgloss.Color("#00ADD8")
	colorNpm   = lipgloss.Color("#CB3837")
	colorPip   = lipgloss.Color("#3776AB")
	colorCargo = lipgloss.Color("#DEA584")
	colorBun   = lipgloss.Color("#FBF0DF")

	// Footer
	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	helpPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Foreground(colorText).
			Padding(1, 2)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorTextBright).
			Bold(true).
			Width(14)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	// Detail panel
	detailBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(1, 3)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Width(14)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorTextBright)

	// Spinner
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)

	// Confirmation prompt
	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWarning).
			Padding(0, 1)

	// Scroll indicator
	scrollStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)
