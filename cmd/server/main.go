package main

import (
	"log"
	"net/http"

	"github.com/letov/high_load_sandbox/internal/config"
	"github.com/letov/high_load_sandbox/internal/db"
	"github.com/letov/high_load_sandbox/internal/handler"
	"github.com/letov/high_load_sandbox/internal/repository"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer database.Close()

	userRepo := repository.NewUserRepository(database)
	userHandler := handler.NewUserHandler(userRepo, []byte(cfg.JWTSecret))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", userHandler.Login)
	mux.HandleFunc("POST /user/register", userHandler.Register)
	mux.HandleFunc("GET /user/get/{id}", userHandler.GetByID)

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
