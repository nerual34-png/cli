package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"grind/internal/auth"
	"grind/internal/levels"
	"grind/internal/tui"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show your stats",
	Long: `Display your personal statistics, XP, and level progress.

Shows:
- Current level and XP
- Progress to next level
- Weekly and total stats
- Quest completion history`,
	RunE: runStats,
}

func runStats(cmd *cobra.Command, args []string) error {
	cfg, err := auth.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.IsLoggedIn() {
		fmt.Println(tui.ErrorStyle.Render("Not logged in. Run 'grind' to set up."))
		return nil
	}

	// TODO: Fetch stats from Convex
	// For now, show starter stats

	totalXP := 0
	weeklyXP := 0
	totalQuests := 0
	weeklyQuests := 0

	level := levels.GetLevel(totalXP)
	nextLevel := levels.GetNextLevel(level)

	// Header
	header := fmt.Sprintf("%s · Level %d · %s",
		tui.TitleStyle.Render(strings.ToUpper(cfg.UserName)),
		level.Number,
		tui.LevelStyle.Render(level.Name),
	)

	separator := tui.MutedStyle.Render(strings.Repeat("═", 48))

	// XP bar
	var xpBar string
	if nextLevel != nil {
		progress := levels.LevelProgress(totalXP)
		barWidth := 30
		xpBar = fmt.Sprintf("%s %d / %d XP",
			tui.ProgressBar(int(progress*float64(barWidth)), barWidth, barWidth),
			totalXP,
			nextLevel.MinXP,
		)
	} else {
		xpBar = tui.ProgressBar(30, 30, 30) + " MAX LEVEL"
	}

	// Stats grid
	statsGrid := fmt.Sprintf(`
  total quests     %d
  this week        %d quests · %d XP
  avg quest        %d XP`,
		totalQuests,
		weeklyQuests, weeklyXP,
		0, // avg XP per quest
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		separator,
		"",
		xpBar,
		"",
		separator,
		statsGrid,
		"",
	)

	box := tui.BoxStyle.Width(55).Render(content)
	fmt.Println(box)

	return nil
}
