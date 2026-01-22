package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"grind/internal/auth"
	"grind/internal/tui"
)

var joinCmd = &cobra.Command{
	Use:   "join [invite-code]",
	Short: "Join a group",
	Long: `Join a friend group using an invite code.

Get an invite code from a friend who has already created a group.
Codes are in the format ABC-123.

Examples:
  grind join ABC-123
  grind join abc123    # Case insensitive`,
	Args: cobra.ExactArgs(1),
	RunE: runJoin,
}

func runJoin(cmd *cobra.Command, args []string) error {
	cfg, err := auth.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.IsLoggedIn() {
		fmt.Println(tui.ErrorStyle.Render("Not logged in. Run 'grind' to set up first."))
		return nil
	}

	if cfg.HasGroup() {
		fmt.Println(tui.ErrorStyle.Render("Already in a group: " + cfg.GroupName))
		return nil
	}

	code := strings.ToUpper(strings.TrimSpace(args[0]))

	// Normalize code format (remove dashes, then add back)
	code = strings.ReplaceAll(code, "-", "")
	if len(code) == 6 {
		code = code[:3] + "-" + code[3:]
	}

	// TODO: Validate with Convex backend
	// For now, accept any code and save locally

	fmt.Print(tui.MutedStyle.Render("  joining..."))

	// Simulate API call delay
	// time.Sleep(500 * time.Millisecond)

	// Clear line
	fmt.Print("\r\033[K")

	// Save to config
	cfg.GroupID = "group_" + code
	cfg.GroupName = "Group " + code
	if err := auth.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println(tui.SuccessStyle.Render("✓ joined " + cfg.GroupName))
	fmt.Println()
	fmt.Println(tui.MutedStyle.Render("run 'grind' to start competing!"))

	return nil
}
