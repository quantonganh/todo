package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantonganh/todo"
)

const (
	dsn = "postgres://postgres:postgres@localhost:5432/todo?sslmode=disable"
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.2-alpine3.15",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_DB=todo",
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{
					HostIP:   "0.0.0.0",
					HostPort: "5432",
				},
			},
		},
	}

	resource, err := pool.RunWithOptions(&opts)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		db := pg.Connect(&pg.Options{
			Addr:     "localhost:5432",
			User:     "postgres",
			Password: "postgres",
			Database: "todo",
		})
		return db.Ping(context.Background())
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func newDB(t *testing.T) (*pg.DB, func()) {
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	schema := uuid.NewV4()
	_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA %q", schema))
	require.NoError(t, err)

	sDB, err := sql.Open("postgres", fmt.Sprintf("%s&search_path=%s", dsn, schema))
	require.NoError(t, err)

	driver, err := postgres.WithInstance(sDB, &postgres.Config{})
	require.NoError(t, err)

	m, err := migrate.NewWithDatabaseInstance(
		"file://../db/migrations",
		"postgres", driver)
	require.NoError(t, err)
	err = m.Up()
	require.True(t, err == nil || errors.Is(err, migrate.ErrNoChange))

	pgDB := pg.Connect(&pg.Options{
		Addr:     "localhost:5432",
		User:     "postgres",
		Password: "postgres",
		Database: "todo",
		OnConnect: func(_ context.Context, conn *pg.Conn) error {
			_, err := conn.Exec("SET search_path=?", schema)
			require.NoError(t, err)
			return nil
		},
	})
	return pgDB, func() {
		_, err = m.Close()
		require.NoError(t, err)

		require.NoError(t, sDB.Close())

		_, err = db.Exec(fmt.Sprintf("DROP SCHEMA %q CASCADE", schema))
		require.NoError(t, err)

		require.NoError(t, db.Close())
	}
}

func Test_taskService_Create(t *testing.T) {
	t.Parallel()

	type args struct {
		task *todo.AddTaskRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				task: &todo.AddTaskRequest{
					Description: "check overdue",
				},
			},
		},
		{
			name: "invalid parent id",
			args: args{
				task: &todo.AddTaskRequest{
					Description: "test invalid parent id",
					ParentID:    intP(10),
				},
			},
			wantErr: &todo.Error{
				Code:    todo.ErrInvalid,
				Message: fmt.Sprintf(ErrMsgParentIDInvalid, 10),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, teardown := newDB(t)
			t.Cleanup(teardown)

			ts := NewTaskService(db)
			err := ts.Create(context.Background(), tt.args.task)
			require.Equal(t, err, tt.wantErr)
		})
	}
}

func Test_taskService_GetByID(t *testing.T) {
	t.Parallel()

	db, teardown := newDB(t)
	t.Cleanup(teardown)
	ts := NewTaskService(db)

	ctx := context.Background()
	require.NoError(t, ts.Create(ctx, &todo.AddTaskRequest{
		Description: "first task",
	}))

	type args struct {
		id int
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "found",
			args: args{
				id: 1,
			},
		},
		{
			name: "not found",
			args: args{
				id: 2,
			},
			wantErr: &todo.Error{
				Code: todo.ErrNotFound,
				Err:  todo.ErrTaskNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ts.GetByID(ctx, tt.args.id)
			require.Equal(t, err, tt.wantErr)
		})
	}
}

func Test_taskService_UpdateByID(t *testing.T) {
	t.Parallel()

	db, teardown := newDB(t)
	t.Cleanup(teardown)
	ts := NewTaskService(db)

	ctx := context.Background()
	require.NoError(t, ts.Create(ctx, &todo.AddTaskRequest{
		Description: "parent task",
	}))
	require.NoError(t, ts.Create(ctx, &todo.AddTaskRequest{
		Description: "child task",
		ParentID:    intP(1),
	}))

	type args struct {
		id   int
		task *todo.Task
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "auto complete parent task",
			args: args{
				id: 2,
				task: &todo.Task{
					Status: todo.StatusCompleted,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.UpdateByID(ctx, tt.args.id, tt.args.task)
			require.Equal(t, err, tt.wantErr)

			parentTask, err := ts.GetByID(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, todo.StatusCompleted, parentTask.Status)
		})
	}
}

func intP(id int) *int {
	return &id
}

func Test_taskService_List(t *testing.T) {
	t.Parallel()

	db, teardown := newDB(t)
	t.Cleanup(teardown)
	ts := NewTaskService(db)

	ctx := context.Background()
	now := time.Now()
	require.NoError(t, ts.Create(ctx, &todo.AddTaskRequest{
		Description: "first task",
		DueDate:     &now,
	}))

	tasks, err := ts.List(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(tasks))

	task, err := ts.GetByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, todo.StatusOverdue, task.Status)
}
