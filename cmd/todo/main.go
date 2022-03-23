package main

import (
	"fmt"
	"log"

	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"

	"github.com/quantonganh/todo"
	"github.com/quantonganh/todo/http"
	"github.com/quantonganh/todo/postgres"
)

func main() {
	err := godotenv.Load("./cmd/todo/.env")
	if err != nil {
		log.Fatal(err)
	}

	var cfg todo.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	db := pg.Connect(&pg.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort),
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
	})
	defer db.Close()

	taskSvc := postgres.NewTaskService(db)
	s := http.NewServer(taskSvc)
	if err := s.Run(cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
