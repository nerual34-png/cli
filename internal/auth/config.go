package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config holds the user's local configuration
type Config struct {
	UserID      string `json:"userId,omitempty"`
	UserName    string `json:"userName,omitempty"`
	GroupID     string `json:"groupId,omitempty"`
	GroupName   string `json:"groupName,omitempty"`
	ConvexURL   string `json:"convexUrl,omitempty"`
}

// DefaultConvexURL is the default Convex deployment URL
const DefaultConvexURL = "https://flippant-okapi-339.convex.cloud"

// ErrNotLoggedIn indicates the user hasn't set up their profile
var ErrNotLoggedIn = errors.New("not logged in - run 'grind' to set up")

// ErrNoGroup indicates the user hasn't joined a group
var ErrNoGroup = errors.New("not in a group - run 'grind join <code>' to join one")

// configDir returns the config directory path
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".grind"), nil
}

// configPath returns the config file path
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config from disk
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the config to disk
func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// IsLoggedIn returns true if the user has set up their profile
func (c *Config) IsLoggedIn() bool {
	return c.UserID != "" && c.UserName != ""
}

// HasGroup returns true if the user is in a group
func (c *Config) HasGroup() bool {
	return c.GroupID != ""
}

// GetConvexURL returns the Convex URL, using default if not set
func (c *Config) GetConvexURL() string {
	if c.ConvexURL != "" {
		return c.ConvexURL
	}
	return DefaultConvexURL
}

// Clear removes all stored credentials
func Clear() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}
