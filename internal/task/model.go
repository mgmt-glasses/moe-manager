package task

import "time"

type Status string

const (
	StatusTodo Status = "todo"
	StatusDone Status = "done"
)

type Task struct {
	ID          string
	UserID      string
	Title       string
	Status      Status
	CreatedAt   time.Time
	CompletedAt *time.Time
}
