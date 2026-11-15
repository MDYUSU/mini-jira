package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MDYUSU/mini-jira/docs"
	"github.com/MDYUSU/mini-jira/internal/auth"
	"github.com/MDYUSU/mini-jira/internal/config"
	"github.com/MDYUSU/mini-jira/internal/database"
	"github.com/MDYUSU/mini-jira/internal/handler"
	"github.com/MDYUSU/mini-jira/internal/repository"
	"github.com/MDYUSU/mini-jira/internal/service"
)

func main() {
	cfg := config.Load()

	if len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET must be set in .env and be at least 32 characters")
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("connected to postgres")

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.TokenTTL)

	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	labelRepo := repository.NewLabelRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	authService := service.NewAuthService(userRepo, tokens)
	userService := service.NewUserService(userRepo)
	orgService := service.NewOrganizationService(orgRepo, projectRepo)
	projectService := service.NewProjectService(projectRepo, orgRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo)
	labelService := service.NewLabelService(labelRepo, taskRepo, projectRepo)
	commentService := service.NewCommentService(commentRepo, taskRepo, projectRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	orgHandler := handler.NewOrganizationHandler(orgService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)
	labelHandler := handler.NewLabelHandler(labelService)
	commentHandler := handler.NewCommentHandler(commentService)

	root := http.NewServeMux()

	docs.Register(root)

	root.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
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

	root.HandleFunc("POST /auth/register", authHandler.Register)
	root.HandleFunc("POST /auth/login", authHandler.Login)

	api := http.NewServeMux()

	api.HandleFunc("GET /auth/me", authHandler.Me)

	api.HandleFunc("GET /users", userHandler.List)
	api.HandleFunc("POST /users", userHandler.Create)
	api.HandleFunc("GET /users/{id}", userHandler.Get)
	api.HandleFunc("PATCH /users/{id}", userHandler.Update)
	api.HandleFunc("DELETE /users/{id}", userHandler.Delete)
	api.HandleFunc("PUT /users/{id}/profile", userHandler.UpsertProfile)

	api.HandleFunc("POST /orgs", orgHandler.Create)
	api.HandleFunc("GET /orgs/{id}", orgHandler.Get)
	api.HandleFunc("GET /orgs/{id}/projects", orgHandler.ListProjects)
	api.HandleFunc("POST /orgs/{id}/projects", projectHandler.Create)

	api.HandleFunc("GET /projects/{id}", projectHandler.Get)
	api.HandleFunc("PATCH /projects/{id}", projectHandler.Update)
	api.HandleFunc("DELETE /projects/{id}", projectHandler.Delete)
	api.HandleFunc("GET /projects/{id}/stats", projectHandler.Stats)
	api.HandleFunc("GET /projects/{id}/workload", projectHandler.Workload)

	api.HandleFunc("POST /projects/{id}/members", projectHandler.AddMember)
	api.HandleFunc("PATCH /projects/{id}/members/{userId}", projectHandler.UpdateMemberRole)
	api.HandleFunc("DELETE /projects/{id}/members/{userId}", projectHandler.RemoveMember)

	api.HandleFunc("GET /projects/{id}/tasks", taskHandler.ListByProject)
	api.HandleFunc("POST /projects/{id}/tasks", taskHandler.Create)
	api.HandleFunc("GET /tasks/{id}", taskHandler.Get)
	api.HandleFunc("PATCH /tasks/{id}", taskHandler.Update)
	api.HandleFunc("DELETE /tasks/{id}", taskHandler.Delete)
	api.HandleFunc("GET /tasks/{id}/activity", taskHandler.ListActivity)
	api.HandleFunc("GET /tasks/{id}/tree", taskHandler.Tree)

	api.HandleFunc("GET /projects/{id}/labels", labelHandler.ListByProject)
	api.HandleFunc("POST /projects/{id}/labels", labelHandler.Create)
	api.HandleFunc("DELETE /labels/{id}", labelHandler.Delete)
	api.HandleFunc("POST /tasks/{id}/labels", labelHandler.Attach)
	api.HandleFunc("DELETE /tasks/{id}/labels/{labelId}", labelHandler.Detach)

	api.HandleFunc("GET /tasks/{id}/comments", commentHandler.ListByTask)
	api.HandleFunc("POST /tasks/{id}/comments", commentHandler.Create)
	api.HandleFunc("PATCH /comments/{id}", commentHandler.Update)
	api.HandleFunc("DELETE /comments/{id}", commentHandler.Delete)

	root.Handle("/", handler.RequireAuth(tokens, api))

	addr := ":" + cfg.AppPort
	log.Printf("server listening on %s", addr)
	log.Printf("api docs at http://localhost%s/docs", addr)
	log.Fatal(http.ListenAndServe(addr, root))
}
