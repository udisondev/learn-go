package user

import "time"

//go:generate go-enum --sql

// SubPlan represents user subscription plan
// ENUM(free, basic, standard, premium)
type SubPlan int

// User represents a user in the system
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Phone        *string
	RegisteredAt time.Time
	UpdatedAt    time.Time
	SubPlan      SubPlan
	Score        int
	IsVerified   bool
	AvatarURL    *string
}
