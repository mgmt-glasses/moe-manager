package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"

	appauth "github.com/mgmt-glasses/moe-manager/internal/auth"
	"github.com/mgmt-glasses/moe-manager/internal/character"
	charadapter "github.com/mgmt-glasses/moe-manager/internal/character/adapter"
	"github.com/mgmt-glasses/moe-manager/internal/chat"
	chatadapter "github.com/mgmt-glasses/moe-manager/internal/chat/adapter"
	"github.com/mgmt-glasses/moe-manager/internal/screentime"
	screentimeadapter "github.com/mgmt-glasses/moe-manager/internal/screentime/adapter"
	"github.com/mgmt-glasses/moe-manager/internal/statistics"
	statsadapter "github.com/mgmt-glasses/moe-manager/internal/statistics/adapter"
	"github.com/mgmt-glasses/moe-manager/internal/task"
	taskadapter "github.com/mgmt-glasses/moe-manager/internal/task/adapter"
	"github.com/mgmt-glasses/moe-manager/internal/user"
	useradapter "github.com/mgmt-glasses/moe-manager/internal/user/adapter"
	"github.com/mgmt-glasses/moe-manager/internal/voice"
	voiceadapter "github.com/mgmt-glasses/moe-manager/internal/voice/adapter"
	migrations "github.com/mgmt-glasses/moe-manager/migrations"
)

var _ user.CharacterValidator = (*charadapter.PostgresCharacterRepository)(nil)

func main() {
	ctx := context.Background()

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

	screentimeRepo := screentimeadapter.NewPostgresScreenTimeRepository(db)
	var screentimeAnalyzer screentime.ImageAnalyzer
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey != "" {
		analyzer, err := screentimeadapter.NewGeminiImageAnalyzer(ctx, geminiKey)
		if err != nil {
			log.Printf("failed to init gemini analyzer, continuing without it: %v", err)
		} else {
			screentimeAnalyzer = analyzer
			defer analyzer.Close()
		}
	} else {
		log.Println("GEMINI_API_KEY is empty, image analysis feature will be disabled")
	}
	screentimeSvc := screentime.NewService(screentimeRepo, screentimeAnalyzer)
	screentimeHandler := screentime.NewHandler(screentimeSvc)

	vertexProject := os.Getenv("VERTEX_PROJECT")
	if vertexProject == "" {
		log.Fatalf("VERTEX_PROJECT is required")
	}
	vertexLocation := os.Getenv("VERTEX_LOCATION")
	if vertexLocation == "" {
		vertexLocation = "us-central1"
	}
	vertexModel := os.Getenv("VERTEX_MODEL")
	if vertexModel == "" {
		vertexModel = "gemini-1.5-flash-001"
	}

	llmClient, err := chatadapter.NewVertexAILLMClient(ctx, vertexProject, vertexLocation, vertexModel)
	if err != nil {
		log.Fatalf("failed to create vertex ai client: %v", err)
	}

	chatContextLoader := chatadapter.NewPostgresChatContextLoader(db)
	personaRepo := chatadapter.NewPostgresPersonaRepository(db)
	chatSvc := chat.NewService(llmClient, chatContextLoader, personaRepo, nil)
	chatHandler := chat.NewHandler(chatSvc)

	ttsURL := os.Getenv("TTS_SERVICE_URL")
	if ttsURL == "" {
		ttsURL = "http://localhost:8001"
	}
	voiceAudioDir := os.Getenv("VOICE_AUDIO_DIR")
	if voiceAudioDir == "" {
		voiceAudioDir = "data/voices"
	}
	// TTS_TIMEOUT_SECONDS: TTS のコールドスタート対策で timeout を可変にする（未設定/不正は 150 秒）。
	ttsTimeout := 150 * time.Second
	if v := os.Getenv("TTS_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttsTimeout = time.Duration(n) * time.Second
		} else {
			log.Printf("invalid TTS_TIMEOUT_SECONDS %q, using default %s", v, ttsTimeout)
		}
	}
	voiceSynthesizer := voiceadapter.NewTTSHTTPClient(ttsURL, ttsTimeout)
	voiceStorage := voiceadapter.NewLocalAudioStorage(voiceAudioDir)
	voiceRepo := voiceadapter.NewPostgresVoiceFileRepository(db)
	voiceUserChecker := voiceadapter.NewPostgresUserCharacterChecker(db)
	voiceSvc := voice.NewService(voiceSynthesizer, voiceStorage, voiceRepo, voiceUserChecker)
	voiceHandler := voice.NewHandler(voiceSvc)

	var authVerifier appauth.TokenVerifier
	if os.Getenv("AUTH_BYPASS") == "true" {
		// ローカル開発・テスト専用。Firebase ログイン未実装のフロントから
		// Authorization: Bearer <userId> だけで認証必須 API を叩けるようにする。
		// 本番では AUTH_BYPASS を設定しないこと（フェイルクローズ）。
		log.Print("WARNING: AUTH_BYPASS=true — 認証バイパス有効。トークン署名を検証しません。ローカル開発専用です。")
		authVerifier = appauth.BypassVerifier{}
	} else {
		firebaseProjectID := os.Getenv("FIREBASE_PROJECT_ID")
		verifier, err := appauth.NewFirebaseVerifier(firebaseProjectID)
		if err != nil {
			log.Fatalf("failed to init auth verifier: %v", err)
		}
		authVerifier = verifier
	}
	authMiddleware := appauth.NewMiddleware(authVerifier)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/api/v1/characters", charHandler.List)
	r.Get("/api/v1/characters/{characterId}", charHandler.GetByID)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)

		r.Post("/api/v1/users", userHandler.Create)

		r.Route("/api/v1/users/{userId}", func(r chi.Router) {
			r.Use(authMiddleware.RequirePathUser("userId"))

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

			r.Post("/screentime/{date}", screentimeHandler.Upsert)
			r.Get("/screentime/{date}", screentimeHandler.Get)
			r.Get("/screentime", screentimeHandler.List)
			r.Post("/screentime/analyze", screentimeHandler.Analyze)

			r.Post("/chat/messages", chatHandler.HandleSendMessage)
			r.Post("/voices", voiceHandler.Generate)
			r.Get("/voice-files/{voiceFileId}", voiceHandler.GetAudio)
		})
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
