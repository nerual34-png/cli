package levels

// Level represents a level in the XP system
type Level struct {
	Number int
	Name   string
	MinXP  int
}

// All levels in the system
var Levels = []Level{
	{Number: 1, Name: "Script Kiddie", MinXP: 0},
	{Number: 2, Name: "Debugger", MinXP: 100},
	{Number: 3, Name: "Builder", MinXP: 300},
	{Number: 4, Name: "Architect", MinXP: 600},
	{Number: 5, Name: "Phantom", MinXP: 1000},
	{Number: 6, Name: "Shadow Dev", MinXP: 1500},
	{Number: 7, Name: "Zero Day", MinXP: 2200},
	{Number: 8, Name: "Root", MinXP: 3000},
	{Number: 9, Name: "God Mode", MinXP: 4000},
	{Number: 10, Name: "∞", MinXP: 5500},
}

// GetLevel returns the level for a given XP amount
func GetLevel(xp int) Level {
	level := Levels[0]
	for _, l := range Levels {
		if xp >= l.MinXP {
			level = l
		} else {
			break
		}
	}
	return level
}

// GetLevelByNumber returns the level by its number
func GetLevelByNumber(num int) Level {
	if num < 1 || num > len(Levels) {
		return Levels[0]
	}
	return Levels[num-1]
}

// GetNextLevel returns the next level after the current one
func GetNextLevel(current Level) *Level {
	if current.Number >= len(Levels) {
		return nil
	}
	return &Levels[current.Number]
}

// XPToNextLevel returns XP needed to reach the next level
func XPToNextLevel(currentXP int) int {
	current := GetLevel(currentXP)
	next := GetNextLevel(current)
	if next == nil {
		return 0
	}
	return next.MinXP - currentXP
}

// LevelProgress returns progress (0.0-1.0) through current level
func LevelProgress(currentXP int) float64 {
	current := GetLevel(currentXP)
	next := GetNextLevel(current)
	if next == nil {
		return 1.0
	}
	levelXP := currentXP - current.MinXP
	levelRange := next.MinXP - current.MinXP
	if levelRange == 0 {
		return 1.0
	}
	return float64(levelXP) / float64(levelRange)
}
