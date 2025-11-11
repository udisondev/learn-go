package exercise

import "time"

//go:generate go-enum --sql

// ExerciseType represents the type of exercise
// ENUM(find_bug, implement_function, complete_code)
type ExerciseType int

// Difficulty represents exercise difficulty level
// ENUM(easy, medium, hard)
type Difficulty int

// TestCase represents a test case for an exercise
type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

// Exercise represents a coding exercise
type Exercise struct {
	ID           int64
	LessonID     int64
	Title        string
	Description  string
	ExerciseType ExerciseType
	StarterCode  string
	TestCases    []TestCase
	Points       int
	Difficulty   Difficulty
	TimeLimit    int // seconds
	MemoryLimit  int // MB
	Order        int
	CreatedAt    time.Time
}
