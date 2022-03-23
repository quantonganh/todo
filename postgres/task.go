package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-pg/pg/v10"

	"github.com/quantonganh/todo"
)

const (
	ErrMsgParentIDInvalid = "parent_id=%d is not present in table"
)

type taskService struct {
	DB *pg.DB
}

func NewTaskService(db *pg.DB) *taskService {
	return &taskService{
		DB: db,
	}
}

// Create creates new task in the database
// Returns ErrInvalid code if parent id is not present in the table
func (ts *taskService) Create(ctx context.Context, task *todo.AddTaskRequest) error {
	query, args, err := sq.
		Insert("task").
		Columns(
			"description",
			"due_date",
			"parent_id",
		).
		Values(
			task.Description,
			task.DueDate,
			task.ParentID,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = ts.DB.ExecContext(ctx, query, args...)
	if err != nil {
		var pgErr pg.Error
		if errors.As(err, &pgErr) {
			if pgErr.IntegrityViolation() {
				return &todo.Error{
					Code:    todo.ErrInvalid,
					Message: fmt.Sprintf(ErrMsgParentIDInvalid, *task.ParentID),
				}
			}
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetByID gets task (and all sub-tasks) by its id
// Returns ErrNotFound if id does not exist
func (ts *taskService) GetByID(ctx context.Context, id int) (*todo.Task, error) {
	query, args, err := sq.
		Select(
			"id",
			"description",
			"status",
			"due_date",
			"subtasks(id) AS sub_tasks",
		).
		From("task").
		Where(sq.Eq{
			"id": id,
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var task todo.Task
	_, err = ts.DB.QueryOneContext(ctx, &task, query, args...)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return nil, &todo.Error{
				Err:  todo.ErrTaskNotFound,
				Code: todo.ErrNotFound,
			}
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &task, nil
}

// UpdateByID updates a task by id
func (ts *taskService) UpdateByID(ctx context.Context, id int, task *todo.Task) error {
	query, args, err := sq.
		Update("task").
		SetMap(map[string]interface{}{
			"description": task.Description,
			"status":      task.Status,
			"due_date":    task.DueDate,
		}).
		Where(sq.Eq{
			"id": id,
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = ts.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// List lists all tasks
func (ts *taskService) List(ctx context.Context) ([]*todo.Task, error) {
	if err := ts.markOverdue(ctx); err != nil {
		return nil, err
	}

	query, args, err := sq.
		Select(
			"id",
			"description",
			"status",
			"due_date",
			"subtasks(id) AS sub_tasks",
		).
		From("task").
		Where(sq.Eq{
			"parent_id": nil,
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	tasks := make([]*todo.Task, 0)
	_, err = ts.DB.QueryContext(ctx, &tasks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return tasks, nil
}

func (ts *taskService) markOverdue(ctx context.Context) error {
	query, args, err := sq.
		Update("task").
		Set("status", todo.StatusOverdue).
		Where(sq.Lt{
			"due_date": time.Now(),
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = ts.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}
