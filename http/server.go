package http

import (
	"fmt"
	"os"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/quantonganh/todo"
)

// Server represents an HTTP server
type Server struct {
	logger zerolog.Logger
	router *gin.Engine

	TaskService todo.TaskService
}

// NewServer creates new HTTP server
func NewServer(taskSvc todo.TaskService) *Server {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	zLogger := log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		},
	)
	log.Logger = zLogger

	s := &Server{
		logger:      zLogger,
		router:      gin.New(),
		TaskService: taskSvc,
	}

	s.router.Use(logger.SetLogger())

	s.router.POST("/tasks", s.Error(s.createTaskHandler()))
	s.router.GET("/tasks/:id", s.Error(s.getTaskHandler()))
	s.router.PATCH("/tasks/:id", s.Error(s.updateTaskHandler()))
	s.router.GET("/tasks", s.Error(s.listTasksHandler()))

	return s
}

// Run starts listening and serving HTTP requests
func (s *Server) Run(port string) error {
	return s.router.Run(fmt.Sprintf(":%s", port))
}
