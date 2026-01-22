package tui

import "github.com/charmbracelet/lipgloss"

// Cyber-HUD Color Palette
var (
	ColorCyberGold   = lipgloss.Color("#FFD700") // Primary/XP
	ColorNeonBlue    = lipgloss.Color("#00BFFF") // Accent/Selection
	ColorMatrixGreen = lipgloss.Color("#04B575") // Success/+XP
	ColorAlertRed    = lipgloss.Color("#FF0055") // Danger/Rivalry
	ColorSlate       = lipgloss.Color("#7D7D7D") // Secondary text
	ColorDimmed      = lipgloss.Color("#404040") // Completed items
	ColorDarkBg      = lipgloss.Color("#1A1A1A") // Panel backgrounds
	ColorPurple      = lipgloss.Color("#BD93F9") // AI/Insight accent

	// Aliases for backward compatibility
	ColorPrimary   = ColorCyberGold
	ColorSecondary = ColorSlate
	ColorSuccess   = ColorMatrixGreen
	ColorAccent    = ColorAlertRed
	ColorMuted     = ColorDimmed
	ColorCyan      = ColorNeonBlue
)

// Styles
var (
	// App container
	AppStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	BoxStyleMuted = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2)

	// Text styles
	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	// XP styles
	XPStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSuccess)

	XPLargeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)

	// Level styles
	LevelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	LevelBadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	// Quest styles
	QuestPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	QuestDoneStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Strikethrough(true)

	QuestSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary)

	// In-progress style - gold/active for working quests
	InProgressStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	// Activity styles
	ActivityStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	ActivityHighlightStyle = lipgloss.NewStyle().
				Foreground(ColorCyan)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1)

	InputFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	// Progress bar
	ProgressFullStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	// Rank styles
	Rank1Style = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary) // Gold

	Rank2Style = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#C0C0C0")) // Silver

	Rank3Style = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#CD7F32")) // Bronze

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// === CYBER-HUD STYLES ===

	// Selection border - thick left border for selected items
	SelectionBorderStyle = lipgloss.NewStyle().
				Foreground(ColorNeonBlue).
				Bold(true)

	// XP Badge - pill style with background
	XPBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(ColorCyberGold).
			Padding(0, 1).
			Bold(true)

	// XP Badge for completed (green)
	XPBadgeCompletedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(ColorMatrixGreen).
				Padding(0, 1).
				Bold(true)

	// Panel title style
	PanelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyberGold)

	// AI Insight box style
	InsightBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorNeonBlue).
			Padding(0, 1)

	// Insight title style
	InsightTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorNeonBlue)

	// Alert/Warning style (for competitive insights)
	AlertStyle = lipgloss.NewStyle().
			Foreground(ColorAlertRed).
			Bold(true)

	// Timestamp style
	TimestampStyle = lipgloss.NewStyle().
			Foreground(ColorSlate)

	// Leaderboard rank styles
	LeaderRankStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyberGold)

	// Streak style
	StreakStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B00")).
			Bold(true)

	// Progress bar bracket style
	ProgressBracketStyle = lipgloss.NewStyle().
				Foreground(ColorSlate)
)

// ProgressBar renders a progress bar
func ProgressBar(current, max, width int) string {
	if max == 0 {
		max = 1
	}
	filled := int(float64(current) / float64(max) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += ProgressFullStyle.Render("█")
	}
	for i := 0; i < empty; i++ {
		bar += ProgressEmptyStyle.Render("░")
	}
	return bar
}

// CenteredBox creates a centered box with content
func CenteredBox(content string, width int) string {
	box := BoxStyle.Width(width).Render(content)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, box)
}

// TitledPanel creates a panel with title in the top border
// Example: ╭─ ⚔️ ACTIVE QUESTS ─────────────╮
func TitledPanel(title, content string, width int, borderColor lipgloss.Color) string {
	titlePart := "╭─ " + title + " "
	titleLen := lipgloss.Width(titlePart)
	remainingWidth := width - titleLen - 1
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Build top border with title
	topBorder := PanelTitleStyle.Foreground(borderColor).Render(titlePart)
	topBorder += lipgloss.NewStyle().Foreground(borderColor).Render(repeat("─", remainingWidth) + "╮")

	// Build sides and content
	lines := splitLines(content)
	var body string
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		padding := width - lineWidth - 4 // 4 = "│ " + " │"
		if padding < 0 {
			padding = 0
		}
		borderStyle := lipgloss.NewStyle().Foreground(borderColor)
		body += borderStyle.Render("│") + " " + line + repeat(" ", padding) + " " + borderStyle.Render("│") + "\n"
	}

	// Build bottom border
	bottomBorder := lipgloss.NewStyle().Foreground(borderColor).Render("╰" + repeat("─", width-2) + "╯")

	return topBorder + "\n" + body + bottomBorder
}

// Helper to repeat a string
func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// Helper to split string into lines
func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	lines = append(lines, current)
	return lines
}

// ProgressBarBracketed renders a progress bar with brackets
// Example: [████████▒▒▒▒▒▒▒▒▒▒▒▒▒▒] 185 / 300 XP
func ProgressBarBracketed(current, max, width int) string {
	if max == 0 {
		max = 1
	}
	filled := int(float64(current) / float64(max) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := ProgressBracketStyle.Render("[")
	for i := 0; i < filled; i++ {
		bar += ProgressFullStyle.Render("█")
	}
	for i := 0; i < empty; i++ {
		bar += ProgressEmptyStyle.Render("▒")
	}
	bar += ProgressBracketStyle.Render("]")

	return bar
}
