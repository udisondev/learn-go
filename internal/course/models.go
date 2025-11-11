package course

import (
	"time"

	"github.com/udisondev/learn-go/internal/user"
)

// Module represents a course module
type Module struct {
	ID              int64
	Title           string
	Description     string
	Order           int
	RequiredScore   int
	RequiredSubPlan user.SubPlan
	CreatedAt       time.Time
}

// Lesson represents a lesson within a module
type Lesson struct {
	ID            int64
	ModuleID      int64
	Title         string
	Order         int
	TheoryContent string
	RequiredScore int
	CreatedAt     time.Time
}
