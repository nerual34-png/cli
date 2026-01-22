package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"grind/internal/api"
	"grind/internal/auth"
)

// Screen represents different screens in the app
type Screen int

const (
	ScreenOnboarding Screen = iota
	ScreenDashboard
	ScreenLeaderboard
	ScreenStats
	ScreenLevelUp
)

// App is the root model for the TUI
type App struct {
	screen       Screen
	config       *auth.Config
	client       *api.Client
	width        int
	height       int
	err          error

	// Screen models
	onboarding   *OnboardingModel
	dashboard    *DashboardModel
	// leaderboard  *LeaderboardModel
	// stats        *StatsModel
}

// NewApp creates a new App instance
func NewApp(cfg *auth.Config) *App {
	var client *api.Client
	if url := cfg.GetConvexURL(); url != "" {
		client = api.NewClient(url)
	}

	app := &App{
		config: cfg,
		client: client,
	}

	// Determine starting screen
	if !cfg.IsLoggedIn() {
		app.screen = ScreenOnboarding
		app.onboarding = NewOnboardingModel(cfg, client)
	} else {
		app.screen = ScreenDashboard
		app.dashboard = NewDashboardModel(cfg, client)
	}

	return app
}

// Init initializes the app
func (a *App) Init() tea.Cmd {
	switch a.screen {
	case ScreenOnboarding:
		return a.onboarding.Init()
	case ScreenDashboard:
		return a.dashboard.Init()
	}
	return nil
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Pass to current screen
		return a.updateCurrentScreen(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q":
			// Only quit on 'q' if not in text input mode
			if a.screen == ScreenOnboarding && a.onboarding != nil && a.onboarding.focusedInput >= 0 {
				// Let the input handle it
			} else if a.screen == ScreenDashboard && a.dashboard != nil && a.dashboard.inputFocused {
				// Let the input handle it
			} else {
				return a, tea.Quit
			}
		}

	case SwitchScreenMsg:
		a.screen = msg.Screen
		switch msg.Screen {
		case ScreenDashboard:
			a.dashboard = NewDashboardModel(a.config, a.client)
			return a, a.dashboard.Init()
		case ScreenOnboarding:
			a.onboarding = NewOnboardingModel(a.config, a.client)
			return a, a.onboarding.Init()
		}
		return a, nil

	case OnboardingCompleteMsg:
		// Save config and switch to dashboard
		a.config = msg.Config
		a.screen = ScreenDashboard
		a.dashboard = NewDashboardModel(a.config, a.client)
		return a, a.dashboard.Init()

	case ErrorMsg:
		a.err = msg.Err
		return a, nil
	}

	return a.updateCurrentScreen(msg)
}

func (a *App) updateCurrentScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch a.screen {
	case ScreenOnboarding:
		if a.onboarding != nil {
			var m tea.Model
			m, cmd = a.onboarding.Update(msg)
			a.onboarding = m.(*OnboardingModel)
		}
	case ScreenDashboard:
		if a.dashboard != nil {
			var m tea.Model
			m, cmd = a.dashboard.Update(msg)
			a.dashboard = m.(*DashboardModel)
		}
	}
	return a, cmd
}

// View renders the app
func (a *App) View() string {
	if a.err != nil {
		return AppStyle.Render(
			ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to quit.", a.err)),
		)
	}

	var content string
	switch a.screen {
	case ScreenOnboarding:
		if a.onboarding != nil {
			content = a.onboarding.View()
		}
	case ScreenDashboard:
		if a.dashboard != nil {
			content = a.dashboard.View()
		}
	default:
		content = "Unknown screen"
	}

	// Apply app styling and center
	return lipgloss.Place(
		a.width,
		a.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// Messages
type SwitchScreenMsg struct {
	Screen Screen
}

type OnboardingCompleteMsg struct {
	Config *auth.Config
}

type ErrorMsg struct {
	Err error
}

// Run starts the TUI application
func Run(cfg *auth.Config) error {
	app := NewApp(cfg)
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
