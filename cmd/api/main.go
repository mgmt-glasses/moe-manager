package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"

	migrations "github.com/mgmt-glasses/moe-manager/migrations"
	"github.com/mgmt-glasses/moe-manager/internal/character"
	charadapter "github.com/mgmt-glasses/moe-manager/internal/character/adapter"
	"github.com/mgmt-glasses/moe-manager/internal/statistics"
	statsadapter "github.com/mgmt-glasses/moe-manager/internal/statistics/adapter"
	"github.com/mgmt-glasses/moe-manager/internal/task"
	taskadapter "github.com/mgmt-glasses/moe-manager/internal/task/adapter"
	"github.com/mgmt-glasses/moe-manager/internal/user"
	useradapter "github.com/mgmt-glasses/moe-manager/internal/user/adapter"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@localhost:5432/moe"
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	charRepo := charadapter.NewPostgresCharacterRepository(db)
	charSvc := character.NewService(charRepo)
	charHandler := character.NewHandler(charSvc)

	userRepo := useradapter.NewPostgresUserRepository(db)
	userSvc := user.NewService(userRepo, charRepo)
	userHandler := user.NewHandler(userSvc)

	statsQuery := statsadapter.NewPostgresStatisticsQuery(db)
	statsSvc := statistics.NewStatisticsService(statsQuery)
	statsHandler := statistics.NewHandler(statsSvc)

	taskRepo := taskadapter.NewPostgresTaskRepository(db)
	taskSvc := task.NewService(taskRepo)
	taskHandler := task.NewHandler(taskSvc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Post("/api/v1/users", userHandler.Create)

	r.Get("/api/v1/characters", charHandler.List)
	r.Get("/api/v1/characters/{characterId}", charHandler.GetByID)

	r.Route("/api/v1/users/{userId}", func(r chi.Router) {
		r.Get("/", userHandler.GetByID)
		r.Patch("/", userHandler.Update)
		r.Patch("/selected-character", userHandler.UpdateSelectedCharacter)

		r.Get("/stats/today", statsHandler.GetToday)
		r.Get("/stats/daily/{date}", statsHandler.GetDaily)
		r.Get("/stats/weekly", statsHandler.GetWeekly)

		r.Post("/tasks", taskHandler.Create)
		r.Get("/tasks", taskHandler.List)
		r.Patch("/tasks/{taskId}/complete", taskHandler.Complete)
		r.Patch("/tasks/{taskId}/reopen", taskHandler.Reopen)
		r.Delete("/tasks/{taskId}", taskHandler.Delete)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func runMigrations(db *sql.DB) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
