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
	"github.com/MDYUSU/mini-jira/internal/service"
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
	taskRepo := repository.NewTaskRepository(db)

	userService := service.NewUserService(userRepo)
	orgService := service.NewOrganizationService(orgRepo, projectRepo)
	projectService := service.NewProjectService(projectRepo, orgRepo)

	userHandler := handler.NewUserHandler(userService)
	orgHandler := handler.NewOrganizationHandler(orgService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskRepo, projectRepo)

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
	mux.HandleFunc("POST /users", userHandler.Create)
	mux.HandleFunc("GET /users/{id}", userHandler.Get)
	mux.HandleFunc("PATCH /users/{id}", userHandler.Update)
	mux.HandleFunc("DELETE /users/{id}", userHandler.Delete)
	mux.HandleFunc("PUT /users/{id}/profile", userHandler.UpsertProfile)

	mux.HandleFunc("POST /orgs", orgHandler.Create)
	mux.HandleFunc("GET /orgs/{id}", orgHandler.Get)
	mux.HandleFunc("GET /orgs/{id}/projects", orgHandler.ListProjects)
	mux.HandleFunc("POST /orgs/{id}/projects", projectHandler.Create)

	mux.HandleFunc("GET /projects/{id}", projectHandler.Get)
	mux.HandleFunc("PATCH /projects/{id}", projectHandler.Update)
	mux.HandleFunc("DELETE /projects/{id}", projectHandler.Delete)
	mux.HandleFunc("GET /projects/{id}/tasks", taskHandler.ListByProject)

	mux.HandleFunc("POST /projects/{id}/members", projectHandler.AddMember)
	mux.HandleFunc("PATCH /projects/{id}/members/{userId}", projectHandler.UpdateMemberRole)
	mux.HandleFunc("DELETE /projects/{id}/members/{userId}", projectHandler.RemoveMember)

	mux.HandleFunc("GET /tasks/{id}", taskHandler.Get)
	mux.HandleFunc("GET /tasks/{id}/activity", taskHandler.ListActivity)

	addr := ":" + cfg.AppPort
	log.Printf("server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}