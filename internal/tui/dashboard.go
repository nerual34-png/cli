package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"grind/internal/api"
	"grind/internal/auth"
	"grind/internal/levels"
	"grind/internal/tui/components"
)

// DashboardModel is the main interactive screen
type DashboardModel struct {
	config       *auth.Config
	client       *api.Client
	width        int
	height       int

	// User data
	user         *api.User
	quests       []api.Quest
	activity     []api.Activity
	leaderboard  []api.LeaderboardEntry
	stats        *api.DashboardStats

	// UI components
	input        textinput.Model
	spinner      spinner.Model
	inputFocused bool
	loading      bool
	err          error

	// Quest selection
	selectedQuest int
	questFocus    bool

	// Cyber-HUD components
	headerComp    *components.HeaderModel
	questPanel    *components.QuestPanelModel
	intelFeed     *components.IntelFeedModel
	animation     *components.AnimationState
	levelUpModal  *components.LevelUpModal
	groupModal    *components.GroupModal
	useCyberHUD   bool // Toggle for new UI
}

// NewDashboardModel creates a new dashboard
func NewDashboardModel(cfg *auth.Config, client *api.Client) *DashboardModel {
	input := textinput.New()
	input.Placeholder = "what's the plan?"
	input.Prompt = "" // Remove default prompt since we add our own
	input.CharLimit = 200
	input.Width = 50
	input.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	// Create mock user from config for now
	user := &api.User{
		ID:       cfg.UserID,
		Name:     cfg.UserName,
		GroupID:  cfg.GroupID,
		TotalXP:  0,
		WeeklyXP: 0,
		Level:    1,
	}

	return &DashboardModel{
		config:        cfg,
		client:        client,
		user:          user,
		quests:        []api.Quest{},
		activity:      []api.Activity{},
		leaderboard:   []api.LeaderboardEntry{},
		input:         input,
		spinner:       s,
		inputFocused:  true,
		selectedQuest: -1,
		// Cyber-HUD components
		headerComp:   components.NewHeader(user, nil, 70),
		questPanel:   components.NewQuestPanel([]api.Quest{}, 36, 14),
		intelFeed:    components.NewIntelFeed([]api.Activity{}, []api.LeaderboardEntry{}, "", cfg.UserName, 38, 14),
		animation:    components.NewAnimationState(),
		levelUpModal: components.NewLevelUpModal(),
		groupModal:   components.NewGroupModal(),
		useCyberHUD:  true, // Enable new UI by default
	}
}

// Init initializes the dashboard
func (d *DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		d.loadUser(),
		d.loadQuests(),
		d.loadActivity(),
		d.loadStats(),
		d.tickActivity(),
	)
}

// loadUser fetches user data from Convex
func (d *DashboardModel) loadUser() tea.Cmd {
	return func() tea.Msg {
		if d.client == nil || d.user.ID == "" {
			return UserLoadedMsg{Err: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := d.client.Query(ctx, "users:get", map[string]any{
			"userId": d.user.ID,
		})
		if err != nil {
			return UserLoadedMsg{Err: err}
		}

		if result == nil {
			return UserLoadedMsg{Err: nil}
		}

		userData, ok := result.(map[string]any)
		if !ok {
			return UserLoadedMsg{Err: nil}
		}

		user := &api.User{
			ID:       userData["_id"].(string),
			Name:     userData["name"].(string),
			TotalXP:  int(userData["totalXp"].(float64)),
			WeeklyXP: int(userData["weeklyXp"].(float64)),
			Level:    int(userData["level"].(float64)),
		}
		if groupId, ok := userData["groupId"].(string); ok {
			user.GroupID = groupId
		}

		return UserLoadedMsg{User: user, Err: nil}
	}
}

// UserLoadedMsg is sent when user data is loaded from Convex
type UserLoadedMsg struct {
	User *api.User
	Err  error
}

// loadActivity fetches activity from Convex
func (d *DashboardModel) loadActivity() tea.Cmd {
	return func() tea.Msg {
		if d.client == nil || d.user.ID == "" {
			return ActivityLoadedMsg{Err: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := d.client.Query(ctx, "activity:getUserActivity", map[string]any{
			"userId": d.user.ID,
			"limit":  20,
		})
		if err != nil {
			return ActivityLoadedMsg{Err: err}
		}

		activitiesData, ok := result.([]any)
		if !ok {
			return ActivityLoadedMsg{Activities: []api.Activity{}, Err: nil}
		}

		var activities []api.Activity
		for _, ad := range activitiesData {
			am, ok := ad.(map[string]any)
			if !ok {
				continue
			}
			activity := api.Activity{
				ID:        am["_id"].(string),
				UserID:    am["userId"].(string),
				Type:      am["type"].(string),
				CreatedAt: int64(am["createdAt"].(float64)),
			}
			if groupId, ok := am["groupId"].(string); ok {
				activity.GroupID = groupId
			}
			if userName, ok := am["userName"].(string); ok {
				activity.UserName = userName
			}
			if questTitle, ok := am["questTitle"].(string); ok {
				activity.QuestTitle = questTitle
			}
			if xp, ok := am["xp"].(float64); ok {
				activity.XP = int(xp)
			}
			if newLevel, ok := am["newLevel"].(float64); ok {
				activity.NewLevel = int(newLevel)
			}
			activities = append(activities, activity)
		}

		return ActivityLoadedMsg{Activities: activities, Err: nil}
	}
}

// ActivityLoadedMsg is sent when activity is loaded from Convex
type ActivityLoadedMsg struct {
	Activities []api.Activity
	Err        error
}

// loadStats fetches dashboard stats from Convex (tries action first, falls back to query)
func (d *DashboardModel) loadStats() tea.Cmd {
	return func() tea.Msg {
		if d.client == nil || d.user.ID == "" {
			return StatsLoadedMsg{Err: nil}
		}

		// Try action first (with AI insight), fall back to query if it fails
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := d.client.Action(ctx, "dashboard:getStatsWithInsight", map[string]any{
			"userId": d.user.ID,
		})

		// If action fails, try the simpler query
		if err != nil {
			ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel2()

			result, err = d.client.Query(ctx2, "dashboard:getStats", map[string]any{
				"userId": d.user.ID,
			})
		}
		if err != nil {
			return StatsLoadedMsg{Err: err}
		}

		if result == nil {
			return StatsLoadedMsg{Err: nil}
		}

		data, ok := result.(map[string]any)
		if !ok {
			return StatsLoadedMsg{Err: nil}
		}

		stats := &api.DashboardStats{}

		// Parse today stats
		if today, ok := data["today"].(map[string]any); ok {
			stats.Today.XP = int(today["xp"].(float64))
			stats.Today.QuestsCompleted = int(today["questsCompleted"].(float64))
			stats.Today.QuestsTotal = int(today["questsTotal"].(float64))
		}

		// Parse week stats
		if week, ok := data["week"].(map[string]any); ok {
			stats.Week.XP = int(week["xp"].(float64))
			stats.Week.Rank = int(week["rank"].(float64))
		}

		// Parse group stats (optional)
		if group, ok := data["group"].(map[string]any); ok {
			stats.Group = &api.GroupStats{
				MemberCount:   int(group["memberCount"].(float64)),
				ActiveToday:   int(group["activeToday"].(float64)),
				UserRank:      int(group["userRank"].(float64)),
				LeaderName:    group["leaderName"].(string),
				LeaderXP:      int(group["leaderXP"].(float64)),
				IsUserLeading: group["isUserLeading"].(bool),
				GroupTodayXP:  int(group["groupTodayXP"].(float64)),
			}
		}

		// Parse quote
		if quote, ok := data["quote"].(string); ok {
			stats.Quote = quote
		}

		// Parse competitive insight (from AI)
		if insight, ok := data["competitiveInsight"].(string); ok {
			stats.CompetitiveInsight = insight
		}

		// Parse insight type for dynamic styling
		if insightType, ok := data["insightType"].(string); ok {
			stats.InsightType = insightType
		}

		return StatsLoadedMsg{Stats: stats, Err: nil}
	}
}

// StatsLoadedMsg is sent when dashboard stats are loaded from Convex
type StatsLoadedMsg struct {
	Stats *api.DashboardStats
	Err   error
}

// loadQuests fetches today's quests from Convex
func (d *DashboardModel) loadQuests() tea.Cmd {
	return func() tea.Msg {
		if d.client == nil {
			return QuestsLoadedMsg{Quests: []api.Quest{}, Err: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := d.client.Query(ctx, "quests:listToday", map[string]any{
			"userId": d.user.ID,
		})
		if err != nil {
			return QuestsLoadedMsg{Err: err}
		}

		// Parse quests from response
		questsData, ok := result.([]any)
		if !ok {
			return QuestsLoadedMsg{Quests: []api.Quest{}, Err: nil}
		}

		var quests []api.Quest
		for _, qd := range questsData {
			qm, ok := qd.(map[string]any)
			if !ok {
				continue
			}
			quest := api.Quest{
				ID:          qm["_id"].(string),
				UserID:      qm["userId"].(string),
				Title:       qm["title"].(string),
				XP:          int(qm["xp"].(float64)),
				AIReasoning: qm["aiReasoning"].(string),
				Status:      qm["status"].(string),
				CreatedAt:   int64(qm["createdAt"].(float64)),
			}
			if groupId, ok := qm["groupId"].(string); ok {
				quest.GroupID = groupId
			}
			if completedAt, ok := qm["completedAt"].(float64); ok {
				quest.CompletedAt = int64(completedAt)
			}
			quests = append(quests, quest)
		}

		return QuestsLoadedMsg{Quests: quests, Err: nil}
	}
}

// tickActivity returns a command that ticks every 5 seconds for activity polling
func (d *DashboardModel) tickActivity() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return ActivityTickMsg{}
	})
}

// ActivityTickMsg is sent when the activity ticker fires
type ActivityTickMsg struct{}

// QuestsLoadedMsg is sent when quests are loaded from Convex
type QuestsLoadedMsg struct {
	Quests []api.Quest
	Err    error
}

// GroupLoadedMsg is sent when group info is loaded
type GroupLoadedMsg struct {
	Name        string
	InviteCode  string
	MemberCount int
	Err         error
}

// loadGroupInfo fetches group info from Convex
func (d *DashboardModel) loadGroupInfo() tea.Cmd {
	return func() tea.Msg {
		if d.client == nil || d.user.GroupID == "" {
			return GroupLoadedMsg{Err: fmt.Errorf("no group")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Get group info
		result, err := d.client.Query(ctx, "groups:get", map[string]any{
			"groupId": d.user.GroupID,
		})
		if err != nil {
			return GroupLoadedMsg{Err: err}
		}

		if result == nil {
			return GroupLoadedMsg{Err: fmt.Errorf("group not found")}
		}

		data, ok := result.(map[string]any)
		if !ok {
			return GroupLoadedMsg{Err: fmt.Errorf("invalid response")}
		}

		name, _ := data["name"].(string)
		inviteCode, _ := data["inviteCode"].(string)

		// Get member count
		membersResult, err := d.client.Query(ctx, "groups:getMembers", map[string]any{
			"groupId": d.user.GroupID,
		})
		memberCount := 0
		if err == nil {
			if members, ok := membersResult.([]any); ok {
				memberCount = len(members)
			}
		}

		return GroupLoadedMsg{
			Name:        name,
			InviteCode:  inviteCode,
			MemberCount: memberCount,
			Err:         nil,
		}
	}
}

// QuestAddedMsg is sent when a quest is added
type QuestAddedMsg struct {
	Quest api.Quest
	Err   error
}

// QuestStartedMsg is sent when a quest is started (pending → in_progress)
type QuestStartedMsg struct {
	QuestID string
	Err     error
}

// QuestCompletedMsg is sent when a quest is completed
type QuestCompletedMsg struct {
	Quest    api.Quest
	XPEarned int
	LevelUp  bool
	NewLevel int
	Err      error
}

// Update handles messages
func (d *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		return d, nil

	case tea.KeyMsg:
		return d.handleKey(msg)

	case ActivityTickMsg:
		// Poll for activity and stats updates
		return d, tea.Batch(d.loadActivity(), d.loadStats(), d.tickActivity())

	case components.AnimationTickMsg:
		// Update animations
		var cmds []tea.Cmd
		if d.animation != nil {
			cmd := d.animation.Update()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if d.levelUpModal != nil {
			cmd := d.levelUpModal.Update()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if len(cmds) > 0 {
			return d, tea.Batch(cmds...)
		}
		return d, nil

	case UserLoadedMsg:
		if msg.Err == nil && msg.User != nil {
			d.user = msg.User
		}
		return d, nil

	case ActivityLoadedMsg:
		if msg.Err == nil && msg.Activities != nil {
			d.activity = msg.Activities
		}
		return d, nil

	case StatsLoadedMsg:
		if msg.Err == nil && msg.Stats != nil {
			d.stats = msg.Stats
		}
		return d, nil

	case QuestsLoadedMsg:
		if msg.Err == nil && msg.Quests != nil {
			d.quests = msg.Quests
		}
		return d, nil

	case GroupLoadedMsg:
		if msg.Err == nil {
			d.groupModal.Show(msg.Name, msg.InviteCode, msg.MemberCount)
		}
		return d, nil

	case QuestAddedMsg:
		d.loading = false
		d.input.SetValue("")
		if msg.Err != nil {
			d.err = msg.Err
			return d, nil
		}
		d.quests = append(d.quests, msg.Quest)
		// Add to activity feed
		d.activity = append([]api.Activity{{
			ID:         fmt.Sprintf("activity_%d", time.Now().UnixNano()),
			UserID:     d.user.ID,
			UserName:   d.user.Name,
			Type:       "quest_created",
			QuestTitle: msg.Quest.Title,
			XP:         msg.Quest.XP,
			CreatedAt:  time.Now().UnixMilli(),
		}}, d.activity...)
		return d, nil

	case QuestStartedMsg:
		if msg.Err != nil {
			d.err = msg.Err
			return d, nil
		}
		// Update quest status locally
		for i := range d.quests {
			if d.quests[i].ID == msg.QuestID {
				d.quests[i].Status = "in_progress"
			}
		}
		return d, nil

	case QuestCompletedMsg:
		if msg.Err != nil {
			d.err = msg.Err
			return d, nil
		}
		// Update quest status
		for i := range d.quests {
			if d.quests[i].ID == msg.Quest.ID {
				d.quests[i].Status = "completed"
				d.quests[i].CompletedAt = time.Now().UnixMilli()
			}
		}
		// Update user XP
		d.user.TotalXP += msg.XPEarned
		d.user.WeeklyXP += msg.XPEarned
		d.user.Level = levels.GetLevel(d.user.TotalXP).Number

		// Add to activity feed
		d.activity = append([]api.Activity{{
			ID:         fmt.Sprintf("activity_%d", time.Now().UnixNano()),
			UserID:     d.user.ID,
			UserName:   d.user.Name,
			Type:       "quest_completed",
			QuestTitle: msg.Quest.Title,
			XP:         msg.XPEarned,
			CreatedAt:  time.Now().UnixMilli(),
		}}, d.activity...) // Prepend to show newest first

		// Trigger quest flash animation
		if d.animation != nil {
			d.animation.TriggerQuestFlash(msg.Quest.ID)
			d.animation.TriggerXPGain(msg.XPEarned, d.user.TotalXP)
		}

		var cmds []tea.Cmd

		if msg.LevelUp {
			// Add level up to activity
			d.activity = append([]api.Activity{{
				ID:        fmt.Sprintf("activity_%d", time.Now().UnixNano()),
				UserID:    d.user.ID,
				UserName:  d.user.Name,
				Type:      "level_up",
				NewLevel:  msg.NewLevel,
				CreatedAt: time.Now().UnixMilli(),
			}}, d.activity...)

			// Show level-up modal
			if d.levelUpModal != nil {
				newLevel := levels.GetLevelByNumber(msg.NewLevel)
				d.levelUpModal.Show(newLevel)
				cmds = append(cmds, components.TickAnimation())
			}
		}

		// Start animation tick if animating
		if d.animation != nil && d.animation.IsAnimating() {
			cmds = append(cmds, components.TickAnimation())
		}

		if len(cmds) > 0 {
			return d, tea.Batch(cmds...)
		}
		return d, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		d.spinner, cmd = d.spinner.Update(msg)
		return d, cmd
	}

	// Update text input
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	return d, cmd
}

func (d *DashboardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Dismiss level-up modal on any keypress
	if d.levelUpModal != nil && d.levelUpModal.Visible {
		d.levelUpModal.Hide()
		return d, nil
	}

	// Dismiss group modal on any keypress
	if d.groupModal != nil && d.groupModal.Visible {
		d.groupModal.Hide()
		return d, nil
	}

	// Clear error on any keypress
	if d.err != nil {
		d.err = nil
	}

	// Global hotkeys (work regardless of input focus)
	switch key {
	case "G":
		// Open group modal - Shift+G
		if d.user.GroupID != "" {
			return d, d.loadGroupInfo()
		} else {
			d.groupModal.ShowNoGroup()
		}
		return d, nil
	}

	// Handle special keys first
	switch key {
	case "enter":
		if d.inputFocused && d.input.Value() != "" {
			return d.addQuest(d.input.Value())
		}
		if d.questFocus && d.selectedQuest >= 0 && d.selectedQuest < len(d.quests) {
			return d.handleQuestAction(d.selectedQuest)
		}
		return d, nil

	case "tab":
		d.inputFocused = !d.inputFocused
		d.questFocus = !d.inputFocused
		if d.inputFocused {
			d.input.Focus()
			d.selectedQuest = -1
			return d, textinput.Blink
		} else {
			d.input.Blur()
			if len(d.quests) > 0 {
				d.selectedQuest = 0
			}
		}
		return d, nil

	case "esc":
		if d.inputFocused {
			d.input.SetValue("")
		}
		return d, nil
	}

	// If input is focused, pass all other keys to the text input
	if d.inputFocused {
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(msg)
		return d, cmd
	}

	// Handle keys when input is NOT focused
	switch key {
	case "up", "k":
		if d.questFocus && d.selectedQuest > 0 {
			d.selectedQuest--
		}
		return d, nil

	case "down", "j":
		if d.questFocus && d.selectedQuest < len(d.quests)-1 {
			d.selectedQuest++
		}
		return d, nil

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx := int(key[0] - '1')
		if idx < len(d.quests) {
			return d.handleQuestAction(idx)
		}

	case "l":
		// TODO: Switch to leaderboard screen

	case "s":
		// TODO: Switch to stats screen

	case "a":
		d.inputFocused = true
		d.questFocus = false
		d.input.Focus()
		d.selectedQuest = -1
		return d, textinput.Blink
	}

	return d, nil
}

func (d *DashboardModel) addQuest(title string) (tea.Model, tea.Cmd) {
	d.loading = true

	return d, func() tea.Msg {
		if d.client == nil {
			// Fallback to local-only mode if no client
			return QuestAddedMsg{Quest: api.Quest{
				ID:          fmt.Sprintf("quest_%d", time.Now().UnixNano()),
				UserID:      d.user.ID,
				GroupID:     d.user.GroupID,
				Title:       title,
				XP:          estimateXP(title),
				AIReasoning: "local mode (no backend)",
				Status:      "pending",
				CreatedAt:   time.Now().UnixMilli(),
			}}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Step 1: Get XP from AI
		var xp int
		var reasoning string

		aiResult, err := d.client.Action(ctx, "ai:evaluateQuest", map[string]any{
			"title": title,
		})
		if err != nil {
			xp = estimateXP(title)
			reasoning = "local estimate"
		} else {
			data, ok := aiResult.(map[string]any)
			if !ok {
				xp = estimateXP(title)
				reasoning = "local estimate"
			} else {
				xp = int(data["xp"].(float64))
				reasoning = data["reasoning"].(string)
			}
		}

		// Step 2: Save quest to Convex
		createResult, err := d.client.Mutation(ctx, "quests:create", map[string]any{
			"userId":      d.user.ID,
			"title":       title,
			"xp":          xp,
			"aiReasoning": reasoning,
		})
		if err != nil {
			return QuestAddedMsg{Err: fmt.Errorf("failed to save quest: %w", err)}
		}

		// Parse the created quest
		data, ok := createResult.(map[string]any)
		if !ok {
			return QuestAddedMsg{Err: fmt.Errorf("invalid response from create")}
		}

		questID, _ := data["questId"].(string)

		return QuestAddedMsg{Quest: api.Quest{
			ID:          questID,
			UserID:      d.user.ID,
			GroupID:     d.user.GroupID,
			Title:       title,
			XP:          xp,
			AIReasoning: reasoning,
			Status:      "pending",
			CreatedAt:   time.Now().UnixMilli(),
		}}
	}
}

// handleQuestAction handles Enter on a quest:
// - pending → in_progress (start)
// - in_progress → completed (complete)
func (d *DashboardModel) handleQuestAction(idx int) (tea.Model, tea.Cmd) {
	if idx < 0 || idx >= len(d.quests) {
		return d, nil
	}
	quest := d.quests[idx]

	switch quest.Status {
	case "pending":
		// Start the quest
		return d, d.startQuest(quest)
	case "in_progress":
		// Complete the quest
		return d, d.completeQuest(quest)
	case "completed":
		// Already done, do nothing
		return d, nil
	}
	return d, nil
}

// startQuest transitions a quest from pending to in_progress
func (d *DashboardModel) startQuest(quest api.Quest) tea.Cmd {
	return func() tea.Msg {
		if d.client == nil {
			// Local-only mode
			return QuestStartedMsg{QuestID: quest.ID}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := d.client.Mutation(ctx, "quests:start", map[string]any{
			"questId": quest.ID,
		})
		if err != nil {
			return QuestStartedMsg{QuestID: quest.ID, Err: err}
		}

		return QuestStartedMsg{QuestID: quest.ID}
	}
}

// completeQuest transitions a quest to completed and earns XP
func (d *DashboardModel) completeQuest(quest api.Quest) tea.Cmd {
	return func() tea.Msg {
		if d.client == nil {
			// Local-only mode
			return QuestCompletedMsg{
				Quest:    quest,
				XPEarned: quest.XP,
				LevelUp:  false,
				NewLevel: 0,
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := d.client.Mutation(ctx, "quests:complete", map[string]any{
			"questId": quest.ID,
		})
		if err != nil {
			return QuestCompletedMsg{Quest: quest, Err: err}
		}

		// Parse response
		data, ok := result.(map[string]any)
		if !ok {
			return QuestCompletedMsg{
				Quest:    quest,
				XPEarned: quest.XP,
				LevelUp:  false,
				NewLevel: 0,
			}
		}

		xpEarned := int(data["xpEarned"].(float64))
		leveledUp, _ := data["leveledUp"].(bool)
		newLevel := 0
		if leveledUp {
			newLevel = int(data["newLevel"].(float64))
		}

		return QuestCompletedMsg{
			Quest:    quest,
			XPEarned: xpEarned,
			LevelUp:  leveledUp,
			NewLevel: newLevel,
		}
	}
}

// estimateXP provides a rough local XP estimate based on task length/keywords
// This is a GRIND app - we reward ACTIVE effort, not passive activities
func estimateXP(title string) int {
	lower := strings.ToLower(title)

	// Passive activities get 0 XP - not a grind task
	passive := []string{"sleep", "rest", "nap", "relax", "chill", "watch", "scroll"}
	for _, kw := range passive {
		if strings.Contains(lower, kw) {
			return 0
		}
	}

	xp := 20 // Base XP for active tasks

	// High effort keywords (+40)
	highEffort := []string{"ship", "deploy", "launch", "build", "implement", "create", "refactor", "marathon", "10km", "20km"}
	for _, kw := range highEffort {
		if strings.Contains(lower, kw) {
			xp += 40
			break
		}
	}

	// Medium effort keywords (+25)
	medEffort := []string{"gym", "workout", "run", "fix", "deep work", "study", "learn", "practice", "write", "design", "code"}
	for _, kw := range medEffort {
		if strings.Contains(lower, kw) {
			xp += 25
			break
		}
	}

	// Small effort keywords (+10)
	smallEffort := []string{"read", "review", "call", "meeting", "email", "update", "check"}
	for _, kw := range smallEffort {
		if strings.Contains(lower, kw) {
			xp += 10
			break
		}
	}

	// Length/complexity bonus
	words := len(strings.Fields(title))
	if words > 5 {
		xp += 10
	}

	// Clamp
	if xp > 100 {
		xp = 100
	}

	return xp
}

// View renders the dashboard
func (d *DashboardModel) View() string {
	// Check for group modal overlay
	if d.groupModal != nil && d.groupModal.Visible {
		return d.groupModal.View(d.width, d.height)
	}

	// Check for level-up modal overlay
	if d.levelUpModal != nil && d.levelUpModal.Visible {
		baseView := d.renderCyberHUD()
		modalView := d.levelUpModal.View(d.width, d.height)
		if modalView != "" {
			return modalView
		}
		return baseView
	}

	if d.useCyberHUD {
		return d.renderCyberHUD()
	}

	return d.renderClassicView()
}

// renderCyberHUD renders the new cyberpunk-style dashboard
func (d *DashboardModel) renderCyberHUD() string {
	// Update component data
	d.headerComp.Update(d.user, d.stats)
	d.questPanel.Update(d.quests, d.selectedQuest, d.questFocus)

	// Get AI insight from stats
	insight := ""
	insightType := ""
	if d.stats != nil {
		insight = d.stats.CompetitiveInsight
		insightType = d.stats.InsightType
	}
	d.intelFeed.Update(d.activity, d.leaderboard, insight, insightType)

	// Render header
	header := d.headerComp.View()

	// Render main panels side by side
	questView := d.questPanel.View()
	intelView := d.intelFeed.View()

	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		questView,
		"  ",
		intelView,
	)

	// Input bar
	inputBar := d.renderInput()

	// Help text
	help := d.renderHelp()

	// Error display
	var errorLine string
	if d.err != nil {
		errorLine = ErrorStyle.Render(fmt.Sprintf("error: %v", d.err))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		mainContent,
		"",
		inputBar,
		help,
		errorLine,
	)
}

// renderClassicView renders the old-style dashboard (fallback)
func (d *DashboardModel) renderClassicView() string {
	// Header with user info
	header := d.renderHeader()

	// Main content: quests and activity side by side
	questPanel := d.renderQuestPanel()
	activityPanel := d.renderActivityPanel()

	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		questPanel,
		"  ",
		activityPanel,
	)

	// Input bar
	inputBar := d.renderInput()

	// Help text
	help := d.renderHelp()

	// Error display
	var errorLine string
	if d.err != nil {
		errorLine = ErrorStyle.Render(fmt.Sprintf("error: %v", d.err))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		mainContent,
		"",
		inputBar,
		help,
		errorLine,
	)
}

func (d *DashboardModel) renderHeader() string {
	level := levels.GetLevelByNumber(d.user.Level)

	// Greeting based on time of day
	hour := time.Now().Hour()
	greeting := "hey"
	if hour < 12 {
		greeting = "gm"
	} else if hour < 17 {
		greeting = "hey"
	} else {
		greeting = "evening"
	}

	// Title line
	title := fmt.Sprintf("%s %s", greeting, d.user.Name)
	levelBadge := LevelBadgeStyle.Render(fmt.Sprintf("L%d %s", level.Number, level.Name))

	titleLine := lipgloss.JoinHorizontal(
		lipgloss.Center,
		TitleStyle.Render(title),
		"  ",
		levelBadge,
	)

	// Stats columns
	var todayCol, weekCol, crewCol string

	if d.stats != nil {
		// Today column
		todayXP := fmt.Sprintf("+%d XP", d.stats.Today.XP)
		if d.stats.Today.XP == 0 {
			todayXP = "0 XP"
		}
		todayQuests := fmt.Sprintf("%d/%d done", d.stats.Today.QuestsCompleted, d.stats.Today.QuestsTotal)
		if d.stats.Today.QuestsTotal == 0 {
			todayQuests = "no quests"
		}
		todayCol = lipgloss.JoinVertical(lipgloss.Left,
			MutedStyle.Render("today"),
			XPStyle.Render(todayXP),
			MutedStyle.Render(todayQuests),
		)

		// Week column
		weekXP := fmt.Sprintf("%d XP", d.stats.Week.XP)
		var weekRank string
		if d.stats.Week.Rank > 0 {
			weekRank = fmt.Sprintf("#%d rank", d.stats.Week.Rank)
		} else {
			weekRank = "no group"
		}
		weekCol = lipgloss.JoinVertical(lipgloss.Left,
			MutedStyle.Render("this week"),
			XPStyle.Render(weekXP),
			MutedStyle.Render(weekRank),
		)

		// Crew column
		if d.stats.Group != nil {
			activeStr := fmt.Sprintf("%d active", d.stats.Group.ActiveToday)
			var leaderStr string
			if d.stats.Group.IsUserLeading {
				leaderStr = "you're leading!"
			} else {
				leaderStr = fmt.Sprintf("%s leading", truncate(d.stats.Group.LeaderName, 10))
			}
			crewCol = lipgloss.JoinVertical(lipgloss.Left,
				MutedStyle.Render("crew"),
				XPStyle.Render(fmt.Sprintf("%d members", d.stats.Group.MemberCount)),
				MutedStyle.Render(activeStr),
				MutedStyle.Render(leaderStr),
			)
		} else {
			crewCol = lipgloss.JoinVertical(lipgloss.Left,
				MutedStyle.Render("crew"),
				MutedStyle.Render("no group"),
				MutedStyle.Render("join one!"),
			)
		}
	} else {
		// Loading state
		todayCol = lipgloss.JoinVertical(lipgloss.Left,
			MutedStyle.Render("today"),
			MutedStyle.Render("..."),
		)
		weekCol = lipgloss.JoinVertical(lipgloss.Left,
			MutedStyle.Render("this week"),
			MutedStyle.Render("..."),
		)
		crewCol = lipgloss.JoinVertical(lipgloss.Left,
			MutedStyle.Render("crew"),
			MutedStyle.Render("..."),
		)
	}

	// Style columns with fixed width
	colStyle := lipgloss.NewStyle().Width(16)
	statsRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		colStyle.Render(todayCol),
		colStyle.Render(weekCol),
		colStyle.Render(crewCol),
	)

	// Competitive insight or quote
	var insightLine string
	if d.stats != nil {
		if d.stats.CompetitiveInsight != "" {
			// AI competitive insight - make it stand out
			insightLine = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Render("→ " + d.stats.CompetitiveInsight)
		} else if d.stats.Quote != "" {
			// Fallback to quote
			insightLine = MutedStyle.Render(fmt.Sprintf("\"%s\"", d.stats.Quote))
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleLine,
		MutedStyle.Render("═══════════════════════════════════════════════"),
		"",
		statsRow,
		"",
		insightLine,
	)

	return BoxStyle.Width(54).Render(content)
}

func (d *DashboardModel) renderQuestPanel() string {
	title := TitleStyle.Render("today's quests")

	// Legend explaining the symbols
	legend := MutedStyle.Render("☐ todo  ◐ working  ✓ done")

	var questLines []string
	activeCount := 0
	potentialXP := 0

	for i, q := range d.quests {
		var line string
		xpStr := XPStyle.Render(fmt.Sprintf("%dXP", q.XP))
		isSelected := d.questFocus && i == d.selectedQuest

		switch q.Status {
		case "completed":
			// ✓ Completed - muted, no XP shown
			if isSelected {
				line = fmt.Sprintf("→  ✓ %s", MutedStyle.Render(truncate(q.Title, 20)))
			} else {
				line = fmt.Sprintf("[%d] ✓ %s", i+1, MutedStyle.Render(truncate(q.Title, 20)))
			}

		case "in_progress":
			// ◐ In progress - highlighted in gold
			activeCount++
			potentialXP += q.XP
			if isSelected {
				line = fmt.Sprintf("→  ◐ %s %s", InProgressStyle.Render(truncate(q.Title, 12)), xpStr)
				line += HelpStyle.Render(" [done]")
			} else {
				line = fmt.Sprintf("[%d] ◐ %s %s", i+1, InProgressStyle.Render(truncate(q.Title, 15)), xpStr)
			}

		default: // "pending"
			// ☐ Pending - normal
			activeCount++
			potentialXP += q.XP
			if isSelected {
				line = fmt.Sprintf("→  ☐ %s %s", QuestSelectedStyle.Render(truncate(q.Title, 12)), xpStr)
				line += HelpStyle.Render(" [start]")
			} else {
				line = fmt.Sprintf("[%d] ☐ %s %s", i+1, truncate(q.Title, 15), xpStr)
			}
		}

		questLines = append(questLines, line)
	}

	if len(questLines) == 0 {
		questLines = append(questLines, MutedStyle.Render("no quests yet"))
		questLines = append(questLines, MutedStyle.Render("type below to add one"))
	}

	// Summary
	var summary string
	if activeCount > 0 {
		summary = fmt.Sprintf("\npotential: %s", XPStyle.Render(fmt.Sprintf("+%d XP", potentialXP)))
	}

	questList := strings.Join(questLines, "\n")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		legend,
		"",
		questList,
		summary,
	)

	return BoxStyleMuted.Width(38).Height(14).Render(content)
}

func (d *DashboardModel) renderActivityPanel() string {
	title := TitleStyle.Render("activity")

	var activityLines []string

	if len(d.activity) == 0 {
		// Show placeholder activity
		activityLines = append(activityLines, MutedStyle.Render("no activity yet"))
		activityLines = append(activityLines, MutedStyle.Render("be the first!"))
	} else {
		// Show up to 8 recent activities
		count := len(d.activity)
		if count > 8 {
			count = 8
		}
		for _, a := range d.activity[:count] {
			var line string
			switch a.Type {
			case "quest_completed":
				line = fmt.Sprintf("✓ %s", truncate(a.QuestTitle, 12))
				activityLines = append(activityLines, SuccessStyle.Render(line))
				activityLines = append(activityLines, XPStyle.Render(fmt.Sprintf("  +%d XP", a.XP)))
			case "quest_started":
				line = fmt.Sprintf("◐ %s", truncate(a.QuestTitle, 12))
				activityLines = append(activityLines, ActivityStyle.Render(line))
			case "quest_created":
				line = fmt.Sprintf("+ %s", truncate(a.QuestTitle, 12))
				activityLines = append(activityLines, ActivityStyle.Render(line))
			case "level_up":
				line = fmt.Sprintf("⚡ LEVEL %d!", a.NewLevel)
				activityLines = append(activityLines, LevelStyle.Render(line))
			default:
				line = fmt.Sprintf("• %s", a.Type)
				activityLines = append(activityLines, ActivityStyle.Render(line))
			}
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		strings.Join(activityLines, "\n"),
	)

	return BoxStyleMuted.Width(22).Height(12).Render(content)
}

// truncate shortens a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func (d *DashboardModel) renderInput() string {
	var prefix string
	if d.loading {
		prefix = d.spinner.View() + " "
	} else {
		prefix = "> "
	}

	style := InputStyle
	if d.inputFocused {
		style = InputFocusedStyle
	}

	// textinput.View() already includes the cursor, just add our prefix
	return style.Width(58).Render(prefix + d.input.View())
}

func (d *DashboardModel) renderHelp() string {
	if d.inputFocused {
		return HelpStyle.Render("enter add task · tab switch to quests · G crew · q quit")
	}
	return HelpStyle.Render("enter start/done · ↑↓ select · G crew · a add · q quit")
}
