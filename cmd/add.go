package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"grind/internal/api"
	"grind/internal/auth"
	"grind/internal/tui"
)

var addCmd = &cobra.Command{
	Use:   "add [task description]",
	Short: "Add a new quest",
	Long: `Add a new quest with AI-evaluated XP.
The AI will assess difficulty and assign points based on:
- Task complexity and scope
- Expected effort and time
- Impact (user-facing vs internal)

Examples:
  grind add "ship landing page"
  grind add "fix auth bug, refactor tests"
  grind add "gym session"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	cfg, err := auth.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.IsLoggedIn() {
		fmt.Println(tui.ErrorStyle.Render("Not logged in. Run 'grind' to set up."))
		return nil
	}

	title := strings.Join(args, " ")

	// Show spinner
	fmt.Print(tui.MutedStyle.Render("  ⠋ evaluating with AI..."))

	// Call Convex AI action to evaluate XP
	xp, reasoning, err := evaluateQuestWithAI(cfg, title)
	if err != nil {
		// Clear spinner and show error
		fmt.Print("\r\033[K")
		fmt.Println(tui.ErrorStyle.Render("AI evaluation failed: " + err.Error()))
		return nil
	}

	// Clear spinner line
	fmt.Print("\r\033[K")

	// Show result
	box := tui.BoxStyle.Width(50).Render(
		fmt.Sprintf("%s · %s\n%s",
			tui.XPStyle.Render(fmt.Sprintf("+%d XP", xp)),
			title,
			tui.MutedStyle.Render("└─ "+reasoning),
		),
	)
	fmt.Println(box)
	fmt.Println(tui.MutedStyle.Render("\nquest added. grind on."))

	return nil
}

// evaluateQuestWithAI calls the Convex AI action to evaluate XP
func evaluateQuestWithAI(cfg *auth.Config, title string) (int, string, error) {
	convexURL := cfg.GetConvexURL()
	if convexURL == "" {
		return 0, "", fmt.Errorf("Convex URL not configured")
	}

	client := api.NewClient(convexURL)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.Action(ctx, "ai:evaluateQuest", map[string]any{
		"title": title,
	})
	if err != nil {
		return 0, "", err
	}

	// Parse response
	data, ok := result.(map[string]any)
	if !ok {
		return 0, "", fmt.Errorf("unexpected response format")
	}

	xp := int(data["xp"].(float64))
	reasoning := data["reasoning"].(string)

	return xp, reasoning, nil
}

// evaluateQuestXP provides local XP estimation
func evaluateQuestXP(title string) (int, string) {
	lower := strings.ToLower(title)
	xp := 25 // Base XP
	reasoning := "standard task"

	// High effort keywords
	highEffort := map[string]string{
		"ship":      "shipping feature, significant effort",
		"deploy":    "deployment, medium-high effort",
		"launch":    "launch, high impact",
		"build":     "building new functionality",
		"implement": "implementation work",
		"create":    "creating new feature",
		"refactor":  "refactoring, medium effort",
		"architect": "architecture work, high complexity",
	}

	// Medium effort keywords
	medEffort := map[string]string{
		"fix":    "bug fix, focused work",
		"update": "update task, moderate effort",
		"review": "code review, careful attention",
		"test":   "testing work",
		"write":  "writing task",
		"design": "design work",
		"debug":  "debugging session",
	}

	// Low effort keywords
	lowEffort := map[string]string{
		"call":    "communication task",
		"email":   "quick communication",
		"meeting": "meeting attendance",
		"read":    "reading task",
		"check":   "quick check",
	}

	// Physical activity
	physicalActivity := map[string]string{
		"gym":     "physical training",
		"workout": "exercise session",
		"run":     "cardio exercise",
		"walk":    "light activity",
	}

	for kw, desc := range highEffort {
		if strings.Contains(lower, kw) {
			xp += 40
			reasoning = desc
			break
		}
	}

	for kw, desc := range medEffort {
		if strings.Contains(lower, kw) {
			xp += 20
			if reasoning == "standard task" {
				reasoning = desc
			}
			break
		}
	}

	for kw, desc := range lowEffort {
		if strings.Contains(lower, kw) {
			xp -= 5
			if reasoning == "standard task" {
				reasoning = desc
			}
			break
		}
	}

	for kw, desc := range physicalActivity {
		if strings.Contains(lower, kw) {
			xp = 35
			reasoning = desc
			break
		}
	}

	// Length/complexity bonus
	words := len(strings.Fields(title))
	if words > 5 {
		xp += 10
	}

	// Clamp
	if xp < 10 {
		xp = 10
	}
	if xp > 100 {
		xp = 100
	}

	return xp, reasoning
}

func init() {
	// Silence default usage
	_ = lipgloss.NewStyle()
	_ = time.Now()
}
