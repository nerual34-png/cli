package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"grind/internal/auth"
	"grind/internal/tui"
)

var doneCmd = &cobra.Command{
	Use:   "done [quest-number]",
	Short: "Complete a quest",
	Long: `Mark a quest as complete and earn XP.

If no quest number is provided, shows a list of pending quests to choose from.

Examples:
  grind done 1    # Complete quest #1
  grind done      # Show list and pick`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDone,
}

func runDone(cmd *cobra.Command, args []string) error {
	cfg, err := auth.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.IsLoggedIn() {
		fmt.Println(tui.ErrorStyle.Render("Not logged in. Run 'grind' to set up."))
		return nil
	}

	// TODO: Get quests from Convex and complete them
	// For now, show placeholder

	if len(args) == 0 {
		fmt.Println(tui.MutedStyle.Render("No quests to complete. Add some with 'grind add \"task\"'"))
		return nil
	}

	questNum, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid quest number: %s", args[0])
	}

	// Placeholder completion animation
	bar := tui.ProgressFullStyle.Render("████████████████████████████████")
	fmt.Println(bar + " " + tui.SuccessStyle.Render("DONE"))
	fmt.Println()
	fmt.Printf(tui.XPStyle.Render("+%d XP")+" · completed quest #%d\n", 50, questNum)

	return nil
}
