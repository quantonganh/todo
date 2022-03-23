package todo

import (
	"context"
	"time"
)

const (
	StatusNew        status = "new"
	StatusInProgress status = "in-progress"
	StatusCompleted  status = "completed"
	StatusOverdue    status = "overdue"
)

// Task represents a todo task
type Task struct {
	ID          int        `json:"id" db:"id"`
	Description string     `json:"description" db:"description"`
	Status      status     `json:"status" db:"status"`
	DueDate     *time.Time `json:"due_date,omitempty" db:"due_date"`
	ParentID    *int       `json:"parent_id,omitempty" db:"parent_id"`
	SubTasks    []Task     `pg:",array" json:"sub_tasks,omitempty" db:"sub_tasks"`
}

// TaskService is the interface that wraps the CRUD methods
type TaskService interface {
	Create(ctx context.Context, task *AddTaskRequest) error
	GetByID(ctx context.Context, id int) (*Task, error)
	UpdateByID(ctx context.Context, id int, task *Task) error
	List(ctx context.Context) ([]*Task, error)
}

// AddTaskRequest represents a request body when adding a new task
type AddTaskRequest struct {
	Description string     `json:"description" binding:"required"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	ParentID    *int       `json:"parent_id,omitempty"`
}

// TaskParam represents path parameters when getting or updating a task
type TaskParam struct {
	ID int `json:"id" uri:"id"`
}

// EditTaskRequest represents a request body when editing a task
type EditTaskRequest struct {
	Description *string    `json:"description,omitempty"`
	Status      *status    `json:"status,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

type status string

var availableStatuses = map[status]struct{}{
	StatusNew:        {},
	StatusInProgress: {},
	StatusCompleted:  {},
	StatusOverdue:    {},
}

// IsValid checks if a task status is valid or not
func (s status) IsValid() bool {
	_, ok := availableStatuses[s]
	return ok
}
