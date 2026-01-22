# Grind

Competitive task tracking for hackers. Add tasks in natural language, AI evaluates XP fairly, and everyone competes on a shared leaderboard.

## Quick Start

```bash
# Build
make build

# Run interactive mode
./grind

# Quick add a task
./grind add "ship landing page"

# Complete a task
./grind done 1

# Check leaderboard
./grind board
```

## Commands

| Command | Description |
|---------|-------------|
| `grind` | Launch interactive TUI |
| `grind add "task"` | Add a new quest with AI-evaluated XP |
| `grind done [n]` | Complete quest #n |
| `grind ls` | List today's quests |
| `grind board` | Show weekly leaderboard |
| `grind stats` | Show your personal stats |
| `grind join <code>` | Join a friend group |

## XP System

Tasks are evaluated based on:
- **Complexity/scope** of work
- **Impact** (user-facing > internal)
- **Effort** required

XP ranges:
- `10-20` Trivial (quick email, small fix)
- `25-40` Standard (routine tasks)
- `45-70` Medium (new features, deep work)
- `75-100` Large (ship major feature)
- `100+` Epic (launch product)

## Levels

| Level | Name | XP Required |
|-------|------|-------------|
| 1 | Script Kiddie | 0 |
| 2 | Debugger | 100 |
| 3 | Builder | 300 |
| 4 | Architect | 600 |
| 5 | Phantom | 1,000 |
| 6 | Shadow Dev | 1,500 |
| 7 | Zero Day | 2,200 |
| 8 | Root | 3,000 |
| 9 | God Mode | 4,000 |
| 10 | ∞ | 5,500 |

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+ (for Convex backend)

### Build

```bash
make build
```

### Run Convex Backend (Development)

```bash
# Install Convex dependencies
cd convex && npm install

# Start dev server
npx convex dev
```

### Project Structure

```
grind/
├── main.go              # Entry point
├── cmd/                 # CLI commands
│   ├── root.go          # Root command + TUI launch
│   ├── add.go           # Add quest
│   ├── done.go          # Complete quest
│   ├── ls.go            # List quests
│   ├── board.go         # Leaderboard
│   ├── stats.go         # User stats
│   └── join.go          # Join group
├── internal/
│   ├── api/             # Convex HTTP client
│   ├── auth/            # Local config management
│   ├── levels/          # XP/level system
│   └── tui/             # Bubbletea TUI
│       ├── app.go       # Main TUI app
│       ├── styles.go    # Lipgloss styles
│       ├── onboarding.go
│       └── dashboard.go
└── convex/              # Convex backend
    ├── schema.ts        # Database schema
    ├── users.ts         # User operations
    ├── groups.ts        # Group management
    ├── quests.ts        # Quest CRUD
    ├── activity.ts      # Activity feed
    └── ai.ts            # AI XP evaluation
```

## Tech Stack

- **CLI**: Go + Cobra + Bubbletea + Lipgloss
- **Backend**: Convex (real-time database)
- **AI**: Claude API (via Convex action)

## License

MIT
