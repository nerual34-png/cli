package components

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"grind/internal/levels"
)

// Animation colors
var (
	animGold   = lipgloss.Color("#FFD700")
	animCyan   = lipgloss.Color("#00D4FF")
	animGreen  = lipgloss.Color("#04B575")
	animWhite  = lipgloss.Color("#FFFFFF")
	animDimmed = lipgloss.Color("#404040")
)

// Animation styles
var (
	levelUpTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(animCyan)

	levelUpLevelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(animGold)

	levelUpNameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(animWhite)

	levelUpBorderStyle = lipgloss.NewStyle().
				Foreground(animGold)

	levelUpHintStyle = lipgloss.NewStyle().
				Foreground(animDimmed)

	xpGainStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(animGreen)
)

// AnimationTickMsg is sent when animation frame should update
type AnimationTickMsg struct{}

// XPGainedMsg is sent when XP is gained
type XPGainedMsg struct {
	Amount   int
	NewTotal int
}

// LevelUpMsg is sent when user levels up
type LevelUpMsg struct {
	NewLevel levels.Level
}

// QuestFlashMsg is sent when a quest should flash
type QuestFlashMsg struct {
	QuestID string
}

// AnimationState manages all animation states
type AnimationState struct {
	// XP counter animation
	DisplayedXP int
	TargetXP    int
	XPTickRate  int

	// Quest flash animation
	FlashQuestID string
	FlashTicks   int

	// Animation frame counter
	Frame int
}

// NewAnimationState creates a new animation state
func NewAnimationState() *AnimationState {
	return &AnimationState{
		DisplayedXP: 0,
		TargetXP:    0,
		XPTickRate:  5, // XP per tick
		Frame:       0,
	}
}

// SetTargetXP sets the XP to animate towards
func (a *AnimationState) SetTargetXP(target int) {
	a.TargetXP = target
}

// SetDisplayedXP sets the currently displayed XP (used for initial load)
func (a *AnimationState) SetDisplayedXP(xp int) {
	a.DisplayedXP = xp
	a.TargetXP = xp
}

// TriggerXPGain starts an XP gain animation
func (a *AnimationState) TriggerXPGain(amount, newTotal int) {
	a.TargetXP = newTotal
	// Adjust tick rate based on amount
	if amount > 50 {
		a.XPTickRate = 10
	} else if amount > 20 {
		a.XPTickRate = 5
	} else {
		a.XPTickRate = 2
	}
}

// TriggerQuestFlash starts a quest flash animation
func (a *AnimationState) TriggerQuestFlash(questID string) {
	a.FlashQuestID = questID
	a.FlashTicks = 6 // ~300ms at 50ms intervals
}

// IsAnimating returns true if any animation is in progress
func (a *AnimationState) IsAnimating() bool {
	return a.DisplayedXP < a.TargetXP || a.FlashTicks > 0
}

// Update updates the animation state
func (a *AnimationState) Update() tea.Cmd {
	updated := false

	// XP tick animation
	if a.DisplayedXP < a.TargetXP {
		a.DisplayedXP += a.XPTickRate
		if a.DisplayedXP > a.TargetXP {
			a.DisplayedXP = a.TargetXP
		}
		updated = true
	}

	// Flash countdown
	if a.FlashTicks > 0 {
		a.FlashTicks--
		updated = true
	}

	// Increment frame
	a.Frame = (a.Frame + 1) % 100

	if updated {
		return TickAnimation()
	}
	return nil
}

// IsQuestFlashing returns true if the given quest should flash
func (a *AnimationState) IsQuestFlashing(questID string) bool {
	return a.FlashQuestID == questID && a.FlashTicks > 0
}

// GetFlashIntensity returns 0-1 flash intensity for visual effects
func (a *AnimationState) GetFlashIntensity() float64 {
	if a.FlashTicks <= 0 {
		return 0
	}
	return float64(a.FlashTicks) / 6.0
}

// TickAnimation returns a command to tick the animation
func TickAnimation() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return AnimationTickMsg{}
	})
}

// LevelUpModal represents the level-up celebration modal
type LevelUpModal struct {
	Level      levels.Level
	Visible    bool
	Ticks      int
	MaxTicks   int
	AutoDismiss bool
}

// NewLevelUpModal creates a new level-up modal
func NewLevelUpModal() *LevelUpModal {
	return &LevelUpModal{
		Visible:     false,
		MaxTicks:    60, // ~3 seconds at 50ms intervals
		AutoDismiss: true,
	}
}

// Show displays the modal for a level up
func (m *LevelUpModal) Show(level levels.Level) {
	m.Level = level
	m.Visible = true
	m.Ticks = 0
}

// Hide hides the modal
func (m *LevelUpModal) Hide() {
	m.Visible = false
}

// Update updates the modal state
func (m *LevelUpModal) Update() tea.Cmd {
	if !m.Visible {
		return nil
	}

	m.Ticks++
	if m.AutoDismiss && m.Ticks >= m.MaxTicks {
		m.Visible = false
		return nil
	}

	return TickAnimation()
}

// View renders the level-up modal
func (m *LevelUpModal) View(screenWidth, screenHeight int) string {
	if !m.Visible {
		return ""
	}

	// Build modal content
	title := levelUpTitleStyle.Render("⚡ LEVEL UP! ⚡")
	levelNum := levelUpLevelStyle.Render(fmt.Sprintf("Level %d", m.Level.Number))
	levelName := levelUpNameStyle.Render(m.Level.Name)
	hint := levelUpHintStyle.Render("press any key to continue...")

	// Combine content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		title,
		"",
		levelNum,
		levelName,
		"",
		hint,
		"",
	)

	// Create modal box
	modalWidth := 30
	modal := m.renderModalBox(content, modalWidth)

	// Center on screen
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}

// renderModalBox renders the modal with double border
func (m *LevelUpModal) renderModalBox(content string, width int) string {
	// Top border
	topBorder := levelUpBorderStyle.Render("╔")
	for i := 0; i < width-2; i++ {
		topBorder += levelUpBorderStyle.Render("═")
	}
	topBorder += levelUpBorderStyle.Render("╗")

	// Content lines
	lines := splitLines(content)
	var body string
	for _, line := range lines {
		lineLen := lipgloss.Width(line)
		totalPadding := width - lineLen - 2
		leftPad := totalPadding / 2
		rightPad := totalPadding - leftPad
		if leftPad < 0 {
			leftPad = 0
		}
		if rightPad < 0 {
			rightPad = 0
		}

		body += levelUpBorderStyle.Render("║")
		for i := 0; i < leftPad; i++ {
			body += " "
		}
		body += line
		for i := 0; i < rightPad; i++ {
			body += " "
		}
		body += levelUpBorderStyle.Render("║") + "\n"
	}

	// Bottom border
	bottomBorder := levelUpBorderStyle.Render("╚")
	for i := 0; i < width-2; i++ {
		bottomBorder += levelUpBorderStyle.Render("═")
	}
	bottomBorder += levelUpBorderStyle.Render("╝")

	return topBorder + "\n" + body + bottomBorder
}
