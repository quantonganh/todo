package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quantonganh/todo"
)

var (
	ErrMarshalFailed = "failed to marshal: %w"
	ErrDecodeFailed  = "failed to decode: %w"
)

func (s *Server) createTaskHandler() appHandler {
	return func(c *gin.Context) error {
		var req todo.AddTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return err
		}

		if req.DueDate != nil {
			today := time.Now()
			if today.After(*req.DueDate) {
				return &todo.Error{
					Code:    todo.ErrInvalid,
					Message: "due_date must be in the future",
				}
			}
		}

		if err := s.TaskService.Create(context.Background(), &req); err != nil {
			return err
		}

		c.JSON(http.StatusOK, req)
		return nil
	}
}

func (s *Server) getTaskHandler() appHandler {
	return func(c *gin.Context) error {
		var param todo.TaskParam
		if err := c.BindUri(&param); err != nil {
			return err
		}

		task, err := s.TaskService.GetByID(context.Background(), param.ID)
		if err != nil {
			return err
		}

		c.JSON(http.StatusOK, task)
		return nil
	}
}

func (s *Server) updateTaskHandler() appHandler {
	return func(c *gin.Context) error {
		var param todo.TaskParam
		if err := c.BindUri(&param); err != nil {
			return err
		}

		var req todo.EditTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return err
		}
		if req.Status != nil && !req.Status.IsValid() {
			return &todo.Error{
				Code:    todo.ErrInvalid,
				Message: "status must be one of: new, in-progress, completed, overdue",
			}
		}
		body, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf(ErrMarshalFailed, err)
		}

		ctx := context.Background()
		tt, err := s.TaskService.GetByID(ctx, param.ID)
		if err != nil {
			return err
		}

		if req.Status != nil && *req.Status == todo.StatusCompleted {
			if tt.ParentID == nil && !s.areAllSubtasksCompleted(tt.SubTasks) {
				return &todo.Error{
					Code:    todo.ErrInvalid,
					Message: "all subtasks must be completed first",
				}
			}
		}

		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&tt); err != nil {
			return fmt.Errorf(ErrDecodeFailed, err)
		}

		err = s.TaskService.UpdateByID(ctx, param.ID, tt)
		if err != nil {
			return err
		}

		c.JSON(http.StatusOK, tt)
		return nil
	}
}

func (s *Server) areAllSubtasksCompleted(subtasks []todo.Task) bool {
	for i := range subtasks {
		if subtasks[i].Status != todo.StatusCompleted {
			return false
		}
	}
	return true
}

func (s *Server) listTasksHandler() appHandler {
	return func(c *gin.Context) error {
		tasks, err := s.TaskService.List(context.Background())
		if err != nil {
			return err
		}

		c.JSON(http.StatusOK, tasks)
		return nil
	}
}
