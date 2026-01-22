package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"grind/internal/auth"
	"grind/internal/tui"
)

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Show leaderboard",
	Long: `Display the weekly leaderboard for your group.

Shows rankings based on XP earned this week.

Examples:
  grind board           # Show weekly leaderboard
  grind board --all     # Show all-time leaderboard`,
	RunE: runBoard,
}

var boardAllTime bool

func runBoard(cmd *cobra.Command, args []string) error {
	cfg, err := auth.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.IsLoggedIn() {
		fmt.Println(tui.ErrorStyle.Render("Not logged in. Run 'grind' to set up."))
		return nil
	}

	if !cfg.HasGroup() {
		fmt.Println(tui.ErrorStyle.Render("Not in a group. Run 'grind join <code>' to join one."))
		return nil
	}

	// Header
	title := "LEADERBOARD · this week"
	if boardAllTime {
		title = "LEADERBOARD · all time"
	}

	// TODO: Fetch leaderboard from Convex
	// For now, show user's own entry

	// Sample leaderboard
	entries := []struct {
		rank   int
		name   string
		level  int
		xp     int
	}{
		{1, cfg.UserName, 1, 0},
	}

	var rows []string
	for _, e := range entries {
		rankStyle := tui.MutedStyle
		switch e.rank {
		case 1:
			rankStyle = tui.Rank1Style
		case 2:
			rankStyle = tui.Rank2Style
		case 3:
			rankStyle = tui.Rank3Style
		}

		// Progress bar
		barWidth := 20
		bar := tui.ProgressBar(e.xp, 100, barWidth)

		row := fmt.Sprintf("  %s  %-12s L%d  %s  %d XP",
			rankStyle.Render(fmt.Sprintf("#%d", e.rank)),
			e.name,
			e.level,
			bar,
			e.xp,
		)
		rows = append(rows, row)
	}

	separator := tui.MutedStyle.Render(strings.Repeat("═", 50))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		tui.TitleStyle.Render(title),
		separator,
		"",
		strings.Join(rows, "\n"),
		"",
		separator,
		tui.MutedStyle.Render("resets in 7 days"),
	)

	box := tui.BoxStyle.Width(55).Render(content)
	fmt.Println(box)

	return nil
}

func init() {
	boardCmd.Flags().BoolVarP(&boardAllTime, "all", "a", false, "Show all-time leaderboard")
}
