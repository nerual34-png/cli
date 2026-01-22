package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"grind/internal/api"
	"grind/internal/levels"
)

// Header colors (referencing main tui package colors)
var (
	headerGold      = lipgloss.Color("#FFD700")
	headerNeonBlue  = lipgloss.Color("#00BFFF")
	headerGreen     = lipgloss.Color("#04B575")
	headerSlate     = lipgloss.Color("#7D7D7D")
	headerDimmed    = lipgloss.Color("#404040")
	headerOrange    = lipgloss.Color("#FF6B00")
)

// Header styles
var (
	headerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(headerGold)

	headerLevelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(headerNeonBlue)

	headerXPStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(headerGold)

	headerMutedStyle = lipgloss.NewStyle().
				Foreground(headerSlate)

	headerStreakStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(headerOrange)

	headerBorderStyle = lipgloss.NewStyle().
				Foreground(headerGold)

	headerProgressFull = lipgloss.NewStyle().
				Foreground(headerGold)

	headerProgressEmpty = lipgloss.NewStyle().
				Foreground(headerDimmed)
)

// HeaderModel represents the header HUD component
type HeaderModel struct {
	User      *api.User
	Stats     *api.DashboardStats
	Level     levels.Level
	NextLevel *levels.Level
	Width     int
}

// NewHeader creates a new header component
func NewHeader(user *api.User, stats *api.DashboardStats, width int) *HeaderModel {
	level := levels.GetLevelByNumber(user.Level)
	nextLevel := levels.GetNextLevel(level)

	return &HeaderModel{
		User:      user,
		Stats:     stats,
		Level:     level,
		NextLevel: nextLevel,
		Width:     width,
	}
}

// Update updates the header with new data
func (h *HeaderModel) Update(user *api.User, stats *api.DashboardStats) {
	h.User = user
	h.Stats = stats
	if user != nil {
		h.Level = levels.GetLevelByNumber(user.Level)
		h.NextLevel = levels.GetNextLevel(h.Level)
	}
}

// View renders the header HUD
func (h *HeaderModel) View() string {
	if h.User == nil {
		return ""
	}

	width := h.Width
	if width < 60 {
		width = 60
	}

	// Build the header content
	line1 := h.renderLevelLine()
	line2 := h.renderStatsLine()

	// Combine lines
	content := line1 + "\n" + line2

	// Render with titled panel style
	return h.renderPanel("GRIND", content, width)
}

// renderLevelLine renders: ⚡ Lvl 2: DEBUGGER [████▒▒▒▒▒▒] 185 / 300 XP
func (h *HeaderModel) renderLevelLine() string {
	// Level info
	levelInfo := headerLevelStyle.Render(fmt.Sprintf("Lvl %d: %s", h.Level.Number, h.Level.Name))

	// Progress bar
	var progressBar, xpText string
	if h.NextLevel != nil {
		progress := levels.LevelProgress(h.User.TotalXP)
		barWidth := 24
		progressBar = h.renderProgressBar(int(progress*float64(barWidth)), barWidth)
		xpText = headerXPStyle.Render(fmt.Sprintf("%d / %d XP", h.User.TotalXP, h.NextLevel.MinXP))
	} else {
		progressBar = h.renderProgressBar(24, 24) // Full bar
		xpText = headerXPStyle.Render("MAX LEVEL")
	}

	return fmt.Sprintf("  %s          %s %s", levelInfo, progressBar, xpText)
}

// renderStatsLine renders: Rank #1 👑 | 🔥 5 Day Streak | 💀 Crew: 2 Active
func (h *HeaderModel) renderStatsLine() string {
	var parts []string

	// Rank
	if h.Stats != nil && h.Stats.Week.Rank > 0 {
		rankIcon := ""
		if h.Stats.Week.Rank == 1 {
			rankIcon = " 👑"
		}
		parts = append(parts, headerMutedStyle.Render(fmt.Sprintf("   Rank #%d%s", h.Stats.Week.Rank, rankIcon)))
	}

	// Streak (placeholder - could add streak tracking later)
	// parts = append(parts, headerStreakStyle.Render("🔥 5 Day Streak"))

	// Weekly XP
	if h.Stats != nil {
		parts = append(parts, headerMutedStyle.Render(fmt.Sprintf("This Week: %d XP", h.Stats.Week.XP)))
	}

	// Crew status
	if h.Stats != nil && h.Stats.Group != nil {
		crewText := fmt.Sprintf("Crew: %d Active", h.Stats.Group.ActiveToday)
		parts = append(parts, headerMutedStyle.Render(crewText))
	}

	// Join with spacing
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "              "
		}
		result += part
	}

	return result
}

// renderProgressBar renders [████████▒▒▒▒▒▒▒▒▒▒▒▒]
func (h *HeaderModel) renderProgressBar(filled, width int) string {
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := headerMutedStyle.Render("[")
	for i := 0; i < filled; i++ {
		bar += headerProgressFull.Render("█")
	}
	for i := 0; i < empty; i++ {
		bar += headerProgressEmpty.Render("▒")
	}
	bar += headerMutedStyle.Render("]")

	return bar
}

// renderPanel creates the bordered panel
func (h *HeaderModel) renderPanel(title, content string, width int) string {
	// Top border with title
	titlePart := "╭── " + title + " "
	titleLen := len(titlePart)
	remainingWidth := width - titleLen - 1
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	topBorder := headerBorderStyle.Render(titlePart)
	for i := 0; i < remainingWidth; i++ {
		topBorder += headerBorderStyle.Render("─")
	}
	topBorder += headerBorderStyle.Render("╮")

	// Content lines with borders
	lines := splitLines(content)
	var body string
	for _, line := range lines {
		lineLen := lipgloss.Width(line)
		padding := width - lineLen - 4
		if padding < 0 {
			padding = 0
		}
		body += headerBorderStyle.Render("│") + " " + line
		for i := 0; i < padding; i++ {
			body += " "
		}
		body += " " + headerBorderStyle.Render("│") + "\n"
	}

	// Bottom border
	bottomBorder := headerBorderStyle.Render("╰")
	for i := 0; i < width-2; i++ {
		bottomBorder += headerBorderStyle.Render("─")
	}
	bottomBorder += headerBorderStyle.Render("╯")

	return topBorder + "\n" + body + bottomBorder
}

// splitLines splits a string by newlines
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
