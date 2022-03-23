package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/quantonganh/todo"
)

var errorToStatusCode = map[string]int{
	todo.ErrInvalid:  http.StatusBadRequest,
	todo.ErrConflict: http.StatusBadRequest,
	todo.ErrNotFound: http.StatusNotFound,
}

type appHandler func(c *gin.Context) error

func (s *Server) Error(fn appHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := fn(c)
		if err == nil {
			return
		}

		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		clientError := new(todo.Error)
		if errors.As(err, &clientError) {
			c.JSON(errorToStatusCode[clientError.Code], clientError)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Oops! Something went wrong.",
		})
		s.logger.Err(err).Msg("")
	}
}
