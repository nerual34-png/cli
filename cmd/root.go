package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"grind/internal/auth"
	"grind/internal/tui"
)

var (
	// Version is set at build time
	Version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "grind",
	Short: "Competitive task tracking for hackers",
	Long: `Grind is a CLI tool for competitive task tracking between friends.
Add tasks in natural language, AI evaluates XP fairly, and everyone
competes on a shared leaderboard.

Run 'grind' without arguments to enter interactive mode.`,
	RunE: runRoot,
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := auth.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Launch interactive TUI
	return tui.Run(cfg)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(boardCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(joinCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("grind %s\n", Version)
	},
}
