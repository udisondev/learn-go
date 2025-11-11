package achievement

import "time"

// Achievement represents an achievement that users can earn
type Achievement struct {
	ID          int64
	Code        string
	Title       string
	Description string
	IconURL     string
	CreatedAt   time.Time
}

// UserAchievement represents a user's earned achievement
type UserAchievement struct {
	UserID        int64
	AchievementID int64
	EarnedAt      time.Time
}
