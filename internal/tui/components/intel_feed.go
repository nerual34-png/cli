package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"grind/internal/api"
)

// Intel feed colors
var (
	intelGold     = lipgloss.Color("#FFD700")
	intelNeonBlue = lipgloss.Color("#00BFFF")
	intelGreen    = lipgloss.Color("#04B575")
	intelRed      = lipgloss.Color("#FF0055")
	intelSlate    = lipgloss.Color("#7D7D7D")
	intelDimmed   = lipgloss.Color("#404040")
	intelWhite    = lipgloss.Color("#FFFFFF")
	intelCyan     = lipgloss.Color("#00D4FF")
)

// Intel feed styles
var (
	intelBorderStyle = lipgloss.NewStyle().
				Foreground(intelSlate)

	intelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(intelGold)

	intelTimestampStyle = lipgloss.NewStyle().
				Foreground(intelSlate)

	intelUserStyle = lipgloss.NewStyle().
			Foreground(intelWhite)

	intelXPStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(intelGreen)

	intelQuestStyle = lipgloss.NewStyle().
			Foreground(intelDimmed)

	intelLevelUpStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(intelCyan)

	// Insight box styles
	insightBorderStyle = lipgloss.NewStyle().
				Foreground(intelNeonBlue)

	insightTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(intelNeonBlue)

	insightTextStyle = lipgloss.NewStyle().
				Foreground(intelWhite)

	insightWarningStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(intelRed)

	// Leaderboard styles
	leaderTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(intelGold)

	leaderRank1Style = lipgloss.NewStyle().
				Bold(true).
				Foreground(intelGold)

	leaderRank2Style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#C0C0C0"))

	leaderRank3Style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CD7F32"))

	leaderNormalStyle = lipgloss.NewStyle().
				Foreground(intelSlate)
)

// IntelFeedModel represents the intel/activity feed component
type IntelFeedModel struct {
	Activities  []api.Activity
	Leaderboard []api.LeaderboardEntry
	AIInsight   string
	InsightType string // "rivalry", "analyst", or "stoic"
	CurrentUser string
	Width       int
	Height      int
}

// NewIntelFeed creates a new intel feed component
func NewIntelFeed(activities []api.Activity, leaderboard []api.LeaderboardEntry, insight, currentUser string, width, height int) *IntelFeedModel {
	return &IntelFeedModel{
		Activities:  activities,
		Leaderboard: leaderboard,
		AIInsight:   insight,
		CurrentUser: currentUser,
		Width:       width,
		Height:      height,
	}
}

// Update updates the intel feed with new data
func (f *IntelFeedModel) Update(activities []api.Activity, leaderboard []api.LeaderboardEntry, insight string, insightType string) {
	f.Activities = activities
	f.Leaderboard = leaderboard
	f.AIInsight = insight
	f.InsightType = insightType
}

// View renders the intel feed
func (f *IntelFeedModel) View() string {
	width := f.Width
	if width < 36 {
		width = 36
	}

	// Build content sections
	var content string

	// Activity feed (top section)
	content += f.renderActivityFeed(4) // Show 4 recent activities

	// AI Insight box (if available)
	if f.AIInsight != "" {
		content += "\n" + f.renderInsightBox()
	}

	// Mini leaderboard
	content += "\n" + f.renderLeaderboard(3) // Show top 3

	return f.renderPanel("INTEL FEED", content, width)
}

// renderActivityFeed renders recent activity in kill-feed style
func (f *IntelFeedModel) renderActivityFeed(maxItems int) string {
	if len(f.Activities) == 0 {
		return intelBorderStyle.Render("no activity yet")
	}

	var lines string
	count := len(f.Activities)
	if count > maxItems {
		count = maxItems
	}

	for _, activity := range f.Activities[:count] {
		lines += f.renderActivity(activity) + "\n"
	}

	return lines
}

// renderActivity renders a single activity item
func (f *IntelFeedModel) renderActivity(a api.Activity) string {
	// Format timestamp
	t := time.UnixMilli(a.CreatedAt)
	timestamp := intelTimestampStyle.Render(fmt.Sprintf("[%s]", t.Format("15:04")))

	// Get user initials (first 2 chars)
	userName := a.UserName
	if userName == "" {
		userName = "??"
	}

	switch a.Type {
	case "quest_completed":
		line1 := fmt.Sprintf("%s %s +%s",
			timestamp,
			intelUserStyle.Render(userName),
			intelXPStyle.Render(fmt.Sprintf("%d XP", a.XP)))
		line2 := "        " + intelQuestStyle.Render(fmt.Sprintf("\"%s\"", truncateString(a.QuestTitle, 16)))
		return line1 + "\n" + line2

	case "quest_started":
		return fmt.Sprintf("%s %s started",
			timestamp,
			intelUserStyle.Render(userName)) + "\n" +
			"        " + intelQuestStyle.Render(fmt.Sprintf("\"%s\"", truncateString(a.QuestTitle, 16)))

	case "quest_created":
		return fmt.Sprintf("%s %s added quest",
			timestamp,
			intelUserStyle.Render(userName))

	case "level_up":
		return intelLevelUpStyle.Render(fmt.Sprintf("%s %s reached LEVEL %d!",
			timestamp, userName, a.NewLevel))

	case "joined_group":
		return fmt.Sprintf("%s %s joined the crew",
			timestamp,
			intelUserStyle.Render(userName))

	default:
		return fmt.Sprintf("%s %s", timestamp, a.Type)
	}
}

// getInsightStyles returns dynamic styles based on insight type
func (f *IntelFeedModel) getInsightStyles() (borderStyle, titleStyle lipgloss.Style, icon, header string) {
	switch f.InsightType {
	case "rivalry":
		// Red - competitive alert when behind
		borderStyle = lipgloss.NewStyle().Foreground(intelRed)
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(intelRed)
		icon = "⚠"
		header = "RIVALRY ALERT"
	case "analyst":
		// Blue - data-driven analysis
		borderStyle = lipgloss.NewStyle().Foreground(intelNeonBlue)
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(intelNeonBlue)
		icon = "📊"
		header = "SYSTEM ANALYSIS"
	default: // "stoic" or empty
		// Gold - motivational/cyberpunk vibes
		borderStyle = lipgloss.NewStyle().Foreground(intelGold)
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(intelGold)
		icon = "💀"
		header = "GRIND MODE"
	}
	return
}

// wrapText wraps text to fit within maxWidth, returning multiple lines
func wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	var lines []string
	words := []rune(text)
	currentLine := ""

	for _, r := range words {
		test := currentLine + string(r)
		if len(test) > maxWidth && currentLine != "" {
			lines = append(lines, currentLine)
			currentLine = string(r)
		} else {
			currentLine = test
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// renderInsightBox renders the AI insight in a nested box with dynamic styling
func (f *IntelFeedModel) renderInsightBox() string {
	innerWidth := f.Width - 8 // Account for outer panel borders and padding
	if innerWidth < 24 {
		innerWidth = 24
	}

	// Get dynamic styles based on insight type
	borderStyle, titleStyle, icon, header := f.getInsightStyles()

	// Title line: "🤖 GEMINI OS"
	title := "🤖 GEMINI OS"

	// Build the box with dynamic border color
	topBorder := borderStyle.Render("┌─ ") +
		titleStyle.Render(title) +
		borderStyle.Render(" ")

	// Fill remaining top border
	titleLen := lipgloss.Width("┌─ " + title + " ")
	remaining := innerWidth - titleLen
	if remaining < 0 {
		remaining = 0
	}
	for i := 0; i < remaining; i++ {
		topBorder += borderStyle.Render("─")
	}
	topBorder += borderStyle.Render("┐")

	// Header line with icon (e.g., "⚠ RIVALRY ALERT")
	headerLine := borderStyle.Render("│ ") +
		titleStyle.Render(icon+" "+header)

	// Pad header to width
	headerLen := lipgloss.Width(headerLine)
	headerPadding := innerWidth - headerLen + 1
	if headerPadding < 0 {
		headerPadding = 0
	}
	for i := 0; i < headerPadding; i++ {
		headerLine += " "
	}
	headerLine += borderStyle.Render("│")

	// Content - wrap insight text across multiple lines
	insightText := f.AIInsight
	maxLineWidth := innerWidth - 6 // Account for borders and padding

	// Wrap text to multiple lines
	wrappedLines := wrapText(insightText, maxLineWidth)

	// Build content lines with quotes
	var contentLines string
	for i, line := range wrappedLines {
		prefix := " "
		suffix := " "
		if i == 0 {
			prefix = "\""
		}
		if i == len(wrappedLines)-1 {
			suffix = "\""
		}

		contentLine := borderStyle.Render("│ ") +
			insightTextStyle.Render(prefix+line+suffix)

		// Pad content to width
		contentLen := lipgloss.Width(contentLine)
		padding := innerWidth - contentLen + 1
		if padding < 0 {
			padding = 0
		}
		for j := 0; j < padding; j++ {
			contentLine += " "
		}
		contentLine += borderStyle.Render("│")
		contentLines += contentLine + "\n"
	}

	// Remove trailing newline from contentLines
	if len(contentLines) > 0 && contentLines[len(contentLines)-1] == '\n' {
		contentLines = contentLines[:len(contentLines)-1]
	}

	// Bottom border
	bottomBorder := borderStyle.Render("└")
	for i := 0; i < innerWidth-2; i++ {
		bottomBorder += borderStyle.Render("─")
	}
	bottomBorder += borderStyle.Render("┘")

	return topBorder + "\n" + headerLine + "\n" + contentLines + "\n" + bottomBorder
}

// renderLeaderboard renders a mini leaderboard
func (f *IntelFeedModel) renderLeaderboard(maxEntries int) string {
	header := leaderTitleStyle.Render("🏆 LEADERBOARD")

	if len(f.Leaderboard) == 0 {
		return header + "\n" + intelBorderStyle.Render("no rankings yet")
	}

	lines := header + "\n"
	count := len(f.Leaderboard)
	if count > maxEntries {
		count = maxEntries
	}

	for i, entry := range f.Leaderboard[:count] {
		rank := i + 1
		var rankStyle lipgloss.Style

		switch rank {
		case 1:
			rankStyle = leaderRank1Style
		case 2:
			rankStyle = leaderRank2Style
		case 3:
			rankStyle = leaderRank3Style
		default:
			rankStyle = leaderNormalStyle
		}

		// Highlight current user
		name := entry.UserName
		if name == f.CurrentUser {
			name = "You"
		}

		lines += rankStyle.Render(fmt.Sprintf("%d. %s (%d XP)", rank, name, entry.WeeklyXP)) + "\n"
	}

	return lines
}

// renderPanel creates the bordered panel with title
func (f *IntelFeedModel) renderPanel(title, content string, width int) string {
	// Top border with title and icon
	titleWithIcon := "📡 " + title
	titlePart := "╭─ " + titleWithIcon + " "
	titleLen := lipgloss.Width(titlePart)
	remainingWidth := width - titleLen - 1
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	topBorder := intelTitleStyle.Render("╭─ " + titleWithIcon + " ")
	for i := 0; i < remainingWidth; i++ {
		topBorder += intelBorderStyle.Render("─")
	}
	topBorder += intelBorderStyle.Render("╮")

	// Content lines with borders
	lines := splitLines(content)
	var body string
	for _, line := range lines {
		lineLen := lipgloss.Width(line)
		padding := width - lineLen - 4
		if padding < 0 {
			padding = 0
		}
		body += intelBorderStyle.Render("│") + " " + line
		for i := 0; i < padding; i++ {
			body += " "
		}
		body += " " + intelBorderStyle.Render("│") + "\n"
	}

	// Bottom border
	bottomBorder := intelBorderStyle.Render("╰")
	for i := 0; i < width-2; i++ {
		bottomBorder += intelBorderStyle.Render("─")
	}
	bottomBorder += intelBorderStyle.Render("╯")

	return topBorder + "\n" + body + bottomBorder
}
