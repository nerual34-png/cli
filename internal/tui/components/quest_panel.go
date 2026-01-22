package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"grind/internal/api"
)

// Quest panel colors
var (
	questGold     = lipgloss.Color("#FFD700")
	questNeonBlue = lipgloss.Color("#00BFFF")
	questGreen    = lipgloss.Color("#04B575")
	questSlate    = lipgloss.Color("#7D7D7D")
	questDimmed   = lipgloss.Color("#404040")
	questWhite    = lipgloss.Color("#FFFFFF")
)

// Quest panel styles
var (
	questPanelBorderStyle = lipgloss.NewStyle().
				Foreground(questSlate)

	questPanelTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(questGold)

	questSelectionBorder = lipgloss.NewStyle().
				Foreground(questNeonBlue).
				Bold(true)

	questPendingStyle = lipgloss.NewStyle().
				Foreground(questWhite)

	questInProgressStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(questGold)

	questCompletedStyle = lipgloss.NewStyle().
				Foreground(questDimmed).
				Strikethrough(true)

	questXPBadgeStyle = lipgloss.NewStyle().
				Foreground(questGold).
				Bold(true)

	questXPCompletedStyle = lipgloss.NewStyle().
				Foreground(questGreen).
				Bold(true)

	questRewardStyle = lipgloss.NewStyle().
				Foreground(questSlate)
)

// Quest status icons
const (
	IconPending    = "[ ]"
	IconInProgress = "[●]"
	IconCompleted  = "[✔]"
)

// QuestPanelModel represents the quest list component
type QuestPanelModel struct {
	Quests   []api.Quest
	Selected int
	Focused  bool
	Width    int
	Height   int
}

// NewQuestPanel creates a new quest panel component
func NewQuestPanel(quests []api.Quest, width, height int) *QuestPanelModel {
	return &QuestPanelModel{
		Quests:   quests,
		Selected: -1,
		Focused:  false,
		Width:    width,
		Height:   height,
	}
}

// Update updates the quest panel with new data
func (q *QuestPanelModel) Update(quests []api.Quest, selected int, focused bool) {
	q.Quests = quests
	q.Selected = selected
	q.Focused = focused
}

// View renders the quest panel
func (q *QuestPanelModel) View() string {
	width := q.Width
	if width < 34 {
		width = 34
	}

	// Build quest list content
	var content string

	if len(q.Quests) == 0 {
		content = questPanelBorderStyle.Render("no quests yet\n")
		content += questPanelBorderStyle.Render("add one below!")
	} else {
		for i, quest := range q.Quests {
			isSelected := q.Focused && i == q.Selected
			questLine := q.renderQuest(quest, i, isSelected)
			content += questLine + "\n"
		}

		// Add potential XP summary
		potentialXP := q.calculatePotentialXP()
		if potentialXP > 0 {
			content += "\n" + questRewardStyle.Render(fmt.Sprintf("Potential: +%d XP", potentialXP))
		}
	}

	return q.renderPanel("ACTIVE QUESTS", content, width)
}

// renderQuest renders a single quest item
func (q *QuestPanelModel) renderQuest(quest api.Quest, idx int, isSelected bool) string {
	// Selection indicator
	var prefix string
	if isSelected {
		prefix = questSelectionBorder.Render("┃") + " "
	} else {
		prefix = "  "
	}

	// Status icon
	var icon string
	var titleStyle lipgloss.Style
	var xpStyle lipgloss.Style

	switch quest.Status {
	case "pending":
		icon = IconPending
		titleStyle = questPendingStyle
		xpStyle = questXPBadgeStyle
	case "in_progress":
		icon = IconInProgress
		titleStyle = questInProgressStyle
		xpStyle = questXPBadgeStyle
	case "completed":
		icon = IconCompleted
		titleStyle = questCompletedStyle
		xpStyle = questXPCompletedStyle
	default:
		icon = IconPending
		titleStyle = questPendingStyle
		xpStyle = questXPBadgeStyle
	}

	// Title (truncated if needed)
	title := truncateString(quest.Title, 20)
	styledTitle := titleStyle.Render(title)

	// First line: icon + title
	line1 := prefix + icon + " " + styledTitle

	// Second line: XP reward (indented)
	var line2 string
	if quest.Status == "completed" {
		line2 = "      " + xpStyle.Render(fmt.Sprintf("+%d XP", quest.XP))
	} else {
		line2 = "      " + questRewardStyle.Render("Reward: ") + xpStyle.Render(fmt.Sprintf("%d XP", quest.XP))
	}

	// Add action hint if selected
	if isSelected {
		var hint string
		if quest.Status == "pending" {
			hint = questPanelBorderStyle.Render(" [start]")
		} else if quest.Status == "in_progress" {
			hint = questPanelBorderStyle.Render(" [done]")
		}
		line1 += hint
	}

	return line1 + "\n" + line2
}

// calculatePotentialXP calculates XP from incomplete quests
func (q *QuestPanelModel) calculatePotentialXP() int {
	total := 0
	for _, quest := range q.Quests {
		if quest.Status != "completed" {
			total += quest.XP
		}
	}
	return total
}

// renderPanel creates the bordered panel with title
func (q *QuestPanelModel) renderPanel(title, content string, width int) string {
	// Top border with title and icon
	titleWithIcon := "⚔️ " + title
	titlePart := "╭─ " + titleWithIcon + " "
	titleLen := lipgloss.Width(titlePart)
	remainingWidth := width - titleLen - 1
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	topBorder := questPanelTitleStyle.Render("╭─ " + titleWithIcon + " ")
	for i := 0; i < remainingWidth; i++ {
		topBorder += questPanelBorderStyle.Render("─")
	}
	topBorder += questPanelBorderStyle.Render("╮")

	// Content lines with borders
	lines := splitLines(content)
	var body string
	for _, line := range lines {
		lineLen := lipgloss.Width(line)
		padding := width - lineLen - 4
		if padding < 0 {
			padding = 0
		}
		body += questPanelBorderStyle.Render("│") + " " + line
		for i := 0; i < padding; i++ {
			body += " "
		}
		body += " " + questPanelBorderStyle.Render("│") + "\n"
	}

	// Bottom border
	bottomBorder := questPanelBorderStyle.Render("╰")
	for i := 0; i < width-2; i++ {
		bottomBorder += questPanelBorderStyle.Render("─")
	}
	bottomBorder += questPanelBorderStyle.Render("╯")

	return topBorder + "\n" + body + bottomBorder
}

// truncateString truncates a string to max length with ellipsis
func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
