package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MDYUSU/mini-jira/docs"
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
	labelRepo := repository.NewLabelRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	userService := service.NewUserService(userRepo)
	orgService := service.NewOrganizationService(orgRepo, projectRepo)
	projectService := service.NewProjectService(projectRepo, orgRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo)
	labelService := service.NewLabelService(labelRepo, taskRepo, projectRepo)
	commentService := service.NewCommentService(commentRepo, taskRepo, projectRepo)

	userHandler := handler.NewUserHandler(userService)
	orgHandler := handler.NewOrganizationHandler(orgService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)
	labelHandler := handler.NewLabelHandler(labelService)
	commentHandler := handler.NewCommentHandler(commentService)

	mux := http.NewServeMux()

	docs.Register(mux)

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

	mux.HandleFunc("POST /projects/{id}/members", projectHandler.AddMember)
	mux.HandleFunc("PATCH /projects/{id}/members/{userId}", projectHandler.UpdateMemberRole)
	mux.HandleFunc("DELETE /projects/{id}/members/{userId}", projectHandler.RemoveMember)

	mux.HandleFunc("GET /projects/{id}/tasks", taskHandler.ListByProject)
	mux.HandleFunc("POST /projects/{id}/tasks", taskHandler.Create)
	mux.HandleFunc("GET /tasks/{id}", taskHandler.Get)
	mux.HandleFunc("PATCH /tasks/{id}", taskHandler.Update)
	mux.HandleFunc("DELETE /tasks/{id}", taskHandler.Delete)
	mux.HandleFunc("GET /tasks/{id}/activity", taskHandler.ListActivity)

	mux.HandleFunc("GET /projects/{id}/labels", labelHandler.ListByProject)
	mux.HandleFunc("POST /projects/{id}/labels", labelHandler.Create)
	mux.HandleFunc("DELETE /labels/{id}", labelHandler.Delete)
	mux.HandleFunc("POST /tasks/{id}/labels", labelHandler.Attach)
	mux.HandleFunc("DELETE /tasks/{id}/labels/{labelId}", labelHandler.Detach)

	mux.HandleFunc("GET /tasks/{id}/comments", commentHandler.ListByTask)
	mux.HandleFunc("POST /tasks/{id}/comments", commentHandler.Create)
	mux.HandleFunc("PATCH /comments/{id}", commentHandler.Update)
	mux.HandleFunc("DELETE /comments/{id}", commentHandler.Delete)

	addr := ":" + cfg.AppPort
	log.Printf("server listening on %s", addr)
	log.Printf("api docs at http://localhost%s/docs", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}