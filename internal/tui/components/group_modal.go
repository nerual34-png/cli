package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Group modal colors
var (
	groupGold    = lipgloss.Color("#FFD700")
	groupCyan    = lipgloss.Color("#00D4FF")
	groupWhite   = lipgloss.Color("#FFFFFF")
	groupDimmed  = lipgloss.Color("#7D7D7D")
	groupMuted   = lipgloss.Color("#404040")
)

// Group modal styles
var (
	groupModalBorderStyle = lipgloss.NewStyle().
				Foreground(groupGold)

	groupModalTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(groupGold)

	groupModalCodeBoxStyle = lipgloss.NewStyle().
				Foreground(groupCyan)

	groupModalCodeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(groupCyan)

	groupModalTextStyle = lipgloss.NewStyle().
				Foreground(groupWhite)

	groupModalHintStyle = lipgloss.NewStyle().
				Foreground(groupDimmed)

	groupModalCommandStyle = lipgloss.NewStyle().
				Foreground(groupCyan)
)

// GroupModal represents the group management modal
type GroupModal struct {
	Visible     bool
	GroupName   string
	InviteCode  string
	MemberCount int
	HasGroup    bool
}

// NewGroupModal creates a new group modal
func NewGroupModal() *GroupModal {
	return &GroupModal{
		Visible:  false,
		HasGroup: false,
	}
}

// Show displays the modal with group info
func (m *GroupModal) Show(groupName, inviteCode string, memberCount int) {
	m.GroupName = groupName
	m.InviteCode = inviteCode
	m.MemberCount = memberCount
	m.HasGroup = true
	m.Visible = true
}

// ShowNoGroup displays the modal for users without a group
func (m *GroupModal) ShowNoGroup() {
	m.HasGroup = false
	m.Visible = true
}

// Hide hides the modal
func (m *GroupModal) Hide() {
	m.Visible = false
}

// View renders the group modal
func (m *GroupModal) View(screenWidth, screenHeight int) string {
	if !m.Visible {
		return ""
	}

	if m.HasGroup {
		return m.renderWithGroup(screenWidth, screenHeight)
	}
	return m.renderNoGroup(screenWidth, screenHeight)
}

// renderWithGroup renders the modal when user has a group
func (m *GroupModal) renderWithGroup(screenWidth, screenHeight int) string {
	modalWidth := 42

	// Build content
	title := groupModalTitleStyle.Render("👥 YOUR CREW")

	groupLine := groupModalTextStyle.Render(fmt.Sprintf("Group: %s", m.GroupName))
	membersLine := groupModalTextStyle.Render(fmt.Sprintf("Members: %d", m.MemberCount))

	// Inner code box
	codeBox := m.renderCodeBox(m.InviteCode, modalWidth-8)

	shareLine := groupModalHintStyle.Render("Share this code with friends!")
	dismissLine := groupModalHintStyle.Render("press any key to close")

	// Combine content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		title,
		"",
		groupLine,
		membersLine,
		"",
		codeBox,
		"",
		shareLine,
		"",
		dismissLine,
		"",
	)

	// Create modal box
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

// renderNoGroup renders the modal when user has no group
func (m *GroupModal) renderNoGroup(screenWidth, screenHeight int) string {
	modalWidth := 42

	// Build content
	title := groupModalTitleStyle.Render("👥 YOUR CREW")

	noGroupLine := groupModalTextStyle.Render("You're not in a crew yet!")

	createCmd := groupModalHintStyle.Render("Run: ") +
		groupModalCommandStyle.Render("grind group create <name>")
	joinCmd := groupModalHintStyle.Render("Or:  ") +
		groupModalCommandStyle.Render("grind group join <code>")

	dismissLine := groupModalHintStyle.Render("press any key to close")

	// Combine content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		title,
		"",
		noGroupLine,
		"",
		createCmd,
		joinCmd,
		"",
		dismissLine,
		"",
	)

	// Create modal box
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

// renderCodeBox renders the invite code in a highlighted box
func (m *GroupModal) renderCodeBox(code string, width int) string {
	innerWidth := width - 4
	if innerWidth < 20 {
		innerWidth = 20
	}

	// Top border
	topBorder := groupModalCodeBoxStyle.Render("┌─ INVITE CODE ")
	remaining := innerWidth - 15
	if remaining < 0 {
		remaining = 0
	}
	for i := 0; i < remaining; i++ {
		topBorder += groupModalCodeBoxStyle.Render("─")
	}
	topBorder += groupModalCodeBoxStyle.Render("┐")

	// Empty line
	emptyLine := groupModalCodeBoxStyle.Render("│")
	for i := 0; i < innerWidth-2; i++ {
		emptyLine += " "
	}
	emptyLine += groupModalCodeBoxStyle.Render("│")

	// Code line (centered)
	codeText := groupModalCodeStyle.Render(code)
	codeLen := lipgloss.Width(codeText)
	totalPadding := innerWidth - codeLen - 2
	leftPad := totalPadding / 2
	rightPad := totalPadding - leftPad
	if leftPad < 0 {
		leftPad = 0
	}
	if rightPad < 0 {
		rightPad = 0
	}

	codeLine := groupModalCodeBoxStyle.Render("│")
	for i := 0; i < leftPad; i++ {
		codeLine += " "
	}
	codeLine += codeText
	for i := 0; i < rightPad; i++ {
		codeLine += " "
	}
	codeLine += groupModalCodeBoxStyle.Render("│")

	// Bottom border
	bottomBorder := groupModalCodeBoxStyle.Render("└")
	for i := 0; i < innerWidth-2; i++ {
		bottomBorder += groupModalCodeBoxStyle.Render("─")
	}
	bottomBorder += groupModalCodeBoxStyle.Render("┘")

	return topBorder + "\n" + emptyLine + "\n" + codeLine + "\n" + emptyLine + "\n" + bottomBorder
}

// renderModalBox renders the modal with double border
func (m *GroupModal) renderModalBox(content string, width int) string {
	// Top border
	topBorder := groupModalBorderStyle.Render("╔")
	for i := 0; i < width-2; i++ {
		topBorder += groupModalBorderStyle.Render("═")
	}
	topBorder += groupModalBorderStyle.Render("╗")

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

		body += groupModalBorderStyle.Render("║")
		for i := 0; i < leftPad; i++ {
			body += " "
		}
		body += line
		for i := 0; i < rightPad; i++ {
			body += " "
		}
		body += groupModalBorderStyle.Render("║") + "\n"
	}

	// Bottom border
	bottomBorder := groupModalBorderStyle.Render("╚")
	for i := 0; i < width-2; i++ {
		bottomBorder += groupModalBorderStyle.Render("═")
	}
	bottomBorder += groupModalBorderStyle.Render("╝")

	return topBorder + "\n" + body + bottomBorder
}
