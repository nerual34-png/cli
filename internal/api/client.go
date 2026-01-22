package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client wraps the Convex HTTP API
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates a new Convex API client
func NewClient(deploymentURL string) *Client {
	return &Client{
		baseURL:    deploymentURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetToken sets the auth token for API calls
func (c *Client) SetToken(token string) {
	c.token = token
}

// ConvexRequest represents a request to the Convex API
type ConvexRequest struct {
	Path   string         `json:"path"`
	Args   map[string]any `json:"args"`
	Format string         `json:"format,omitempty"`
}

// ConvexResponse represents a response from the Convex API
type ConvexResponse struct {
	Status       string `json:"status"`
	Value        any    `json:"value"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	ErrorData    any    `json:"errorData,omitempty"`
}

// Query executes a Convex query function
func (c *Client) Query(ctx context.Context, path string, args map[string]any) (any, error) {
	return c.call(ctx, "/api/query", path, args)
}

// Mutation executes a Convex mutation function
func (c *Client) Mutation(ctx context.Context, path string, args map[string]any) (any, error) {
	return c.call(ctx, "/api/mutation", path, args)
}

// Action executes a Convex action function
func (c *Client) Action(ctx context.Context, path string, args map[string]any) (any, error) {
	return c.call(ctx, "/api/action", path, args)
}

func (c *Client) call(ctx context.Context, endpoint, path string, args map[string]any) (any, error) {
	if args == nil {
		args = make(map[string]any)
	}

	reqBody := ConvexRequest{
		Path:   path,
		Args:   args,
		Format: "json",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %d: %s", resp.StatusCode, string(respBody))
	}

	var result ConvexResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Status == "error" {
		return nil, fmt.Errorf("convex error: %s", result.ErrorMessage)
	}

	return result.Value, nil
}

// User represents a user in the system
type User struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	GroupID     string `json:"groupId,omitempty"`
	TotalXP     int    `json:"totalXp"`
	WeeklyXP    int    `json:"weeklyXp"`
	Level       int    `json:"level"`
	CreatedAt   int64  `json:"createdAt"`
	LastActiveAt int64 `json:"lastActiveAt"`
}

// Group represents a friend group
type Group struct {
	ID         string `json:"_id"`
	Name       string `json:"name"`
	InviteCode string `json:"inviteCode"`
	CreatedBy  string `json:"createdBy"`
	CreatedAt  int64  `json:"createdAt"`
}

// Quest represents a task/quest
type Quest struct {
	ID          string `json:"_id"`
	UserID      string `json:"userId"`
	GroupID     string `json:"groupId,omitempty"`
	Title       string `json:"title"`
	XP          int    `json:"xp"`
	AIReasoning string `json:"aiReasoning"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"createdAt"`
	CompletedAt int64  `json:"completedAt,omitempty"`
}

// Activity represents an activity feed item
type Activity struct {
	ID         string `json:"_id"`
	GroupID    string `json:"groupId"`
	UserID     string `json:"userId"`
	UserName   string `json:"userName,omitempty"`
	Type       string `json:"type"`
	QuestTitle string `json:"questTitle,omitempty"`
	XP         int    `json:"xp,omitempty"`
	NewLevel   int    `json:"newLevel,omitempty"`
	CreatedAt  int64  `json:"createdAt"`
}

// LeaderboardEntry represents a user's position on the leaderboard
type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
	Level    int    `json:"level"`
	WeeklyXP int    `json:"weeklyXp"`
	TotalXP  int    `json:"totalXp"`
}

// DashboardStats contains aggregated stats for the dashboard header
type DashboardStats struct {
	Today              TodayStats  `json:"today"`
	Week               WeekStats   `json:"week"`
	Group              *GroupStats `json:"group"`
	Quote              string      `json:"quote"`
	CompetitiveInsight string      `json:"competitiveInsight"`
	InsightType        string      `json:"insightType"` // "rivalry", "analyst", or "stoic"
}

// TodayStats contains today's activity stats
type TodayStats struct {
	XP              int `json:"xp"`
	QuestsCompleted int `json:"questsCompleted"`
	QuestsTotal     int `json:"questsTotal"`
}

// WeekStats contains this week's stats
type WeekStats struct {
	XP   int `json:"xp"`
	Rank int `json:"rank"`
}

// GroupStats contains group/crew stats
type GroupStats struct {
	MemberCount   int    `json:"memberCount"`
	ActiveToday   int    `json:"activeToday"`
	UserRank      int    `json:"userRank"`
	LeaderName    string `json:"leaderName"`
	LeaderXP      int    `json:"leaderXP"`
	IsUserLeading bool   `json:"isUserLeading"`
	GroupTodayXP  int    `json:"groupTodayXP"`
}
