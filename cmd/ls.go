package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"grind/internal/auth"
	"grind/internal/tui"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List today's quests",
	Long: `Show all pending and completed quests for today.

Examples:
  grind ls           # List all today's quests
  grind ls --all     # List all quests (not just today)`,
	RunE: runLs,
}

var lsAll bool

func runLs(cmd *cobra.Command, args []string) error {
	cfg, err := auth.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.IsLoggedIn() {
		fmt.Println(tui.ErrorStyle.Render("Not logged in. Run 'grind' to set up."))
		return nil
	}

	// TODO: Fetch quests from Convex
	// For now, show placeholder

	title := "today's quests"
	if lsAll {
		title = "all quests"
	}

	fmt.Println(tui.TitleStyle.Render(title))
	fmt.Println()
	fmt.Println(tui.MutedStyle.Render("  No quests yet. Add some with 'grind add \"task\"'"))
	fmt.Println()

	return nil
}

func init() {
	lsCmd.Flags().BoolVarP(&lsAll, "all", "a", false, "Show all quests, not just today's")
}
