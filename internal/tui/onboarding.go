package tui

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"grind/internal/api"
	"grind/internal/auth"
)

// OnboardingStep represents steps in the onboarding flow
type OnboardingStep int

const (
	StepWelcome OnboardingStep = iota
	StepName
	StepGroupChoice
	StepCreateGroup
	StepJoinGroup
	StepComplete
)

// OnboardingModel handles first-time user setup
type OnboardingModel struct {
	config       *auth.Config
	client       *api.Client
	step         OnboardingStep
	nameInput    textinput.Model
	groupInput   textinput.Model
	codeInput    textinput.Model
	focusedInput int // -1 = no input focused, 0+ = input index
	groupChoice  int // 0 = create, 1 = join
	inviteCode   string
	loading      bool
	err          error
}

// UserCreatedMsg is sent when user is created in Convex
type UserCreatedMsg struct {
	UserID string
	Err    error
}

// GroupCreatedMsg is sent when group is created in Convex
type GroupCreatedMsg struct {
	GroupID    string
	InviteCode string
	Err        error
}

// NewOnboardingModel creates a new onboarding model
func NewOnboardingModel(cfg *auth.Config, client *api.Client) *OnboardingModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "your name"
	nameInput.CharLimit = 32
	nameInput.Width = 30

	groupInput := textinput.New()
	groupInput.Placeholder = "group name"
	groupInput.CharLimit = 32
	groupInput.Width = 30

	codeInput := textinput.New()
	codeInput.Placeholder = "ABC-123"
	codeInput.CharLimit = 10
	codeInput.Width = 15

	return &OnboardingModel{
		config:       cfg,
		client:       client,
		step:         StepWelcome,
		nameInput:    nameInput,
		groupInput:   groupInput,
		codeInput:    codeInput,
		focusedInput: -1,
	}
}

// Init initializes the model
func (m *OnboardingModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *OnboardingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil // Ignore input while loading
		}
		switch msg.String() {
		case "enter":
			return m.handleEnter()
		case "up", "k":
			if m.step == StepGroupChoice {
				m.groupChoice = 0
			}
		case "down", "j":
			if m.step == StepGroupChoice {
				m.groupChoice = 1
			}
		case "esc":
			if m.focusedInput >= 0 {
				m.focusedInput = -1
			}
		}

	case UserCreatedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.config.UserID = msg.UserID
		m.nameInput.Blur()
		m.focusedInput = -1
		m.step = StepGroupChoice
		return m, nil

	case GroupCreatedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.config.GroupID = msg.GroupID
		m.inviteCode = msg.InviteCode
		m.groupInput.Blur()
		m.focusedInput = -1
		m.step = StepComplete
		return m, nil
	}

	// Update text inputs
	var cmd tea.Cmd
	switch m.step {
	case StepName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case StepCreateGroup:
		m.groupInput, cmd = m.groupInput.Update(msg)
	case StepJoinGroup:
		m.codeInput, cmd = m.codeInput.Update(msg)
	}

	return m, cmd
}

func (m *OnboardingModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case StepWelcome:
		m.step = StepName
		m.nameInput.Focus()
		m.focusedInput = 0
		return m, textinput.Blink

	case StepName:
		name := strings.TrimSpace(m.nameInput.Value())
		if name == "" {
			return m, nil
		}
		m.config.UserName = name
		m.loading = true
		m.err = nil

		// Call API to create user
		return m, m.createUserCmd(name)

	case StepGroupChoice:
		if m.groupChoice == 0 {
			m.step = StepCreateGroup
			m.groupInput.Focus()
			m.focusedInput = 0
			return m, textinput.Blink
		} else {
			m.step = StepJoinGroup
			m.codeInput.Focus()
			m.focusedInput = 0
			return m, textinput.Blink
		}

	case StepCreateGroup:
		groupName := strings.TrimSpace(m.groupInput.Value())
		if groupName == "" {
			return m, nil
		}
		m.config.GroupName = groupName
		m.loading = true
		m.err = nil

		// Call API to create group
		return m, m.createGroupCmd(groupName)

	case StepJoinGroup:
		code := strings.TrimSpace(m.codeInput.Value())
		if code == "" {
			return m, nil
		}
		// TODO: Validate invite code with backend
		m.config.GroupID = "joined-group" // Placeholder
		m.config.GroupName = "Joined Group"
		m.codeInput.Blur()
		m.focusedInput = -1
		m.step = StepComplete
		return m, nil

	case StepComplete:
		// Save config and transition
		if err := auth.Save(m.config); err != nil {
			m.err = err
			return m, nil
		}
		return m, func() tea.Msg {
			return OnboardingCompleteMsg{Config: m.config}
		}
	}

	return m, nil
}

// createUserCmd creates a user in Convex
func (m *OnboardingModel) createUserCmd(name string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return UserCreatedMsg{Err: fmt.Errorf("no API client available")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := m.client.Mutation(ctx, "users:create", map[string]any{
			"name": name,
		})
		if err != nil {
			return UserCreatedMsg{Err: err}
		}

		// Result is the user ID string
		userID, ok := result.(string)
		if !ok {
			return UserCreatedMsg{Err: fmt.Errorf("unexpected response type: %T", result)}
		}

		return UserCreatedMsg{UserID: userID}
	}
}

// createGroupCmd creates a group in Convex
func (m *OnboardingModel) createGroupCmd(groupName string) tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			// Fallback to local IDs if no client
			return GroupCreatedMsg{
				GroupID:    generateGroupID(),
				InviteCode: generateInviteCode(),
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := m.client.Mutation(ctx, "groups:create", map[string]any{
			"name":      groupName,
			"createdBy": m.config.UserID,
		})
		if err != nil {
			return GroupCreatedMsg{Err: err}
		}

		// Result should have groupId and inviteCode
		resultMap, ok := result.(map[string]any)
		if !ok {
			return GroupCreatedMsg{Err: fmt.Errorf("unexpected response type: %T", result)}
		}

		groupID, _ := resultMap["groupId"].(string)
		inviteCode, _ := resultMap["inviteCode"].(string)

		return GroupCreatedMsg{
			GroupID:    groupID,
			InviteCode: inviteCode,
		}
	}
}

// View renders the onboarding screen
func (m *OnboardingModel) View() string {
	switch m.step {
	case StepWelcome:
		return m.viewWelcome()
	case StepName:
		return m.viewName()
	case StepGroupChoice:
		return m.viewGroupChoice()
	case StepCreateGroup:
		return m.viewCreateGroup()
	case StepJoinGroup:
		return m.viewJoinGroup()
	case StepComplete:
		return m.viewComplete()
	}
	return ""
}

func (m *OnboardingModel) viewWelcome() string {
	logo := LogoStyle.Render("⚡ GRIND")
	subtitle := SubtitleStyle.Render("competitive task tracking")
	tagline := MutedStyle.Render("for hackers")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		logo,
		"",
		subtitle,
		tagline,
	)

	box := BoxStyle.Width(44).Render(content)
	help := HelpStyle.Render("\npress enter to start")

	return lipgloss.JoinVertical(lipgloss.Center, box, help)
}

func (m *OnboardingModel) viewName() string {
	title := TitleStyle.Render("first time? let's set up.")
	prompt := "\nyour name: " + m.nameInput.View()

	var statusLine string
	if m.loading {
		statusLine = "\n" + MutedStyle.Render("creating account...")
	} else if m.err != nil {
		statusLine = "\n" + ErrorStyle.Render(fmt.Sprintf("error: %v", m.err))
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		prompt,
		statusLine,
	)

	return BoxStyle.Width(44).Render(content)
}

func (m *OnboardingModel) viewGroupChoice() string {
	title := TitleStyle.Render(fmt.Sprintf("hey %s!", m.config.UserName))
	question := "\njoin existing group or create new?"

	create := "  create new group"
	join := "  join with invite code"

	if m.groupChoice == 0 {
		create = QuestSelectedStyle.Render("→ create new group")
	} else {
		join = QuestSelectedStyle.Render("→ join with invite code")
	}

	options := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		create,
		join,
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		question,
		options,
	)

	help := HelpStyle.Render("\n↑/↓ to select, enter to confirm")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		BoxStyle.Width(44).Render(content),
		help,
	)
}

func (m *OnboardingModel) viewCreateGroup() string {
	title := TitleStyle.Render("create your group")
	prompt := "\ngroup name: " + m.groupInput.View()

	var statusLine string
	if m.loading {
		statusLine = "\n" + MutedStyle.Render("creating group...")
	} else if m.err != nil {
		statusLine = "\n" + ErrorStyle.Render(fmt.Sprintf("error: %v", m.err))
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		prompt,
		statusLine,
	)

	return BoxStyle.Width(44).Render(content)
}

func (m *OnboardingModel) viewJoinGroup() string {
	title := TitleStyle.Render("join a group")
	prompt := "\ninvite code: " + m.codeInput.View()

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		prompt,
	)

	return BoxStyle.Width(44).Render(content)
}

func (m *OnboardingModel) viewComplete() string {
	title := SuccessStyle.Render("✓ you're all set!")

	var groupInfo string
	if m.inviteCode != "" {
		groupInfo = fmt.Sprintf("\ninvite your friends:\n\n%s",
			BoxStyleMuted.Render("grind join "+m.inviteCode))
	} else {
		groupInfo = fmt.Sprintf("\njoined: %s", m.config.GroupName)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		groupInfo,
	)

	help := HelpStyle.Render("\npress enter to start grinding...")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		BoxStyle.Width(44).Render(content),
		help,
	)
}

// Helper functions
func generateUserID() string {
	// Simple local ID for now - will be replaced by Convex ID
	return fmt.Sprintf("user_%d", randomID())
}

func generateGroupID() string {
	return fmt.Sprintf("group_%d", randomID())
}

func generateInviteCode() string {
	// Simple invite code format: XXX-XXX
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[randomID()%int64(len(chars))]
	}
	return string(code[:3]) + "-" + string(code[3:])
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomID() int64 {
	return time.Now().UnixNano() + rng.Int63n(10000)
}
