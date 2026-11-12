package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MDYUSU/mini-jira/internal/config"
	"github.com/MDYUSU/mini-jira/internal/database"
	"github.com/MDYUSU/mini-jira/internal/handler"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("connected to postgres")

	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	userHandler := handler.NewUserHandler(userRepo)
	orgHandler := handler.NewOrganizationHandler(orgRepo, projectRepo)
	projectHandler := handler.NewProjectHandler(projectRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status, code := "ok", http.StatusOK
		if err := db.PingContext(ctx); err != nil {
			status, code = "db_unreachable", http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]string{"status": status})
	})

	mux.HandleFunc("GET /users", userHandler.List)
	mux.HandleFunc("GET /users/{id}", userHandler.Get)

	mux.HandleFunc("GET /orgs/{id}", orgHandler.Get)
	mux.HandleFunc("GET /orgs/{id}/projects", orgHandler.ListProjects)

	mux.HandleFunc("GET /projects/{id}", projectHandler.Get)

	addr := ":" + cfg.AppPort
	log.Printf("server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}