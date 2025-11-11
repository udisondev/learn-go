package submission

import "time"

//go:generate go-enum --sql

// SubmissionStatus represents the lifecycle status of a submission
// ENUM(pending, running, completed)
type SubmissionStatus int

// ExecutionStatus represents the result status of code execution
// ENUM(success, failed, error)
type ExecutionStatus int

// Submission represents a user's code submission
type Submission struct {
	ID          int64
	UserID      int64
	ExerciseID  int64
	Code        string
	Status      SubmissionStatus
	SubmittedAt time.Time
}

// TestCaseResult represents the result of a single test case
type TestCaseResult struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Passed   bool   `json:"passed"`
}

// ExecutionResult represents the result of code execution
type ExecutionResult struct {
	ID            int64
	SubmissionID  int64
	Status        ExecutionStatus
	TestResults   []TestCaseResult
	ErrorMessage  *string
	ExecutionTime *int
}

// UserProgress tracks user's progress on exercises
type UserProgress struct {
	UserID        int64
	ExerciseID    int64
	IsCompleted   bool
	Attempts      int
	FirstSolvedAt *time.Time
	UpdatedAt     time.Time
}
