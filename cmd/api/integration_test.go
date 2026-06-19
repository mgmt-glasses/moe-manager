package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// setupTestDB connects to the database configured via DATABASE_URL (falling
// back to the same default as main), applies migrations, and returns the
// connection. If no database is reachable, the test is skipped so these
// integration tests never block local development without Postgres running.
// The connection is closed via t.Cleanup, which runs LIFO: registering it
// here (before any test registers its own row-cleanup) guarantees the DB
// closes last, after row cleanup has had a chance to run.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", resolveDatabaseURL())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.Ping(); err != nil {
		t.Skipf("postgres not reachable, skipping integration test: %v", err)
	}
	if err := runMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	return db
}

// cleanupUser removes all rows created for userID so repeated local test runs
// don't accumulate leftover data in a persistent development database.
func cleanupUser(t *testing.T, db *sql.DB, userID string) {
	t.Helper()
	if _, err := db.Exec(`DELETE FROM tasks WHERE user_id = $1`, userID); err != nil {
		t.Errorf("cleanup tasks: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM users WHERE id = $1`, userID); err != nil {
		t.Errorf("cleanup user: %v", err)
	}
}

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func decodeEnvelope(t *testing.T, body []byte) envelope {
	t.Helper()
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v (body=%s)", err, body)
	}
	return env
}

func doRequest(t *testing.T, h http.Handler, method, path, body string) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, decodeEnvelope(t, rec.Body.Bytes())
}

// TestUserFlow_CreateSelectTaskAndStats walks through the MVP's primary user
// flow against a real PostgreSQL database: create a user, select a secretary
// character, register and complete a task, then confirm both the API
// response and the underlying DB rows reflect the change.
func TestUserFlow_CreateSelectTaskAndStats(t *testing.T) {
	db := setupTestDB(t)
	router := newRouter(db)

	// 1. キャラ一覧から先頭のキャラ ID を取得する(マイグレーションで4キャラ投入済み)。
	rec, env := doRequest(t, router, http.MethodGet, "/api/v1/characters", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /characters status: got %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body)
	}
	var characters []struct {
		CharacterID string `json:"characterId"`
	}
	if err := json.Unmarshal(env.Data, &characters); err != nil {
		t.Fatalf("decode characters: %v", err)
	}
	if len(characters) == 0 {
		t.Fatal("expected seeded characters, got none")
	}
	characterID := characters[0].CharacterID

	// 2. ユーザを作成する。
	rec, env = doRequest(t, router, http.MethodPost, "/api/v1/users",
		`{"name":"統合テスト太郎","presidentName":"統合テスト社長","targetEntertainmentMinutes":90}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /users status: got %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body)
	}
	var createdUser struct {
		UserID string `json:"userId"`
	}
	json.Unmarshal(env.Data, &createdUser)
	if createdUser.UserID == "" {
		t.Fatal("expected userId in response")
	}
	userID := createdUser.UserID
	t.Cleanup(func() { cleanupUser(t, db, userID) })

	// 3. 秘書キャラを選択する。
	rec, _ = doRequest(t, router, http.MethodPatch, "/api/v1/users/"+userID+"/selected-character",
		`{"characterId":"`+characterID+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH selected-character status: got %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body)
	}

	// DB に選択中キャラが反映されているか直接確認する。
	var selectedCharID sql.NullString
	if err := db.QueryRow(`SELECT selected_character_id FROM users WHERE id = $1`, userID).Scan(&selectedCharID); err != nil {
		t.Fatalf("query selected_character_id: %v", err)
	}
	if !selectedCharID.Valid || selectedCharID.String != characterID {
		t.Errorf("selected_character_id in DB: got %v, want %q", selectedCharID, characterID)
	}

	// 4. タスクを登録する。
	rec, env = doRequest(t, router, http.MethodPost, "/api/v1/users/"+userID+"/tasks", `{"title":"資料作成"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /tasks status: got %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body)
	}
	var createdTask struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	json.Unmarshal(env.Data, &createdTask)
	if createdTask.Status != "todo" {
		t.Errorf("task status: got %q, want todo", createdTask.Status)
	}

	// DB にタスクが期待通り保存されているか直接確認する。
	var dbTitle, dbStatus string
	if err := db.QueryRow(`SELECT title, status FROM tasks WHERE id = $1 AND user_id = $2`, createdTask.ID, userID).
		Scan(&dbTitle, &dbStatus); err != nil {
		t.Fatalf("query task: %v", err)
	}
	if dbTitle != "資料作成" || dbStatus != "todo" {
		t.Errorf("task in DB: got title=%q status=%q, want title=資料作成 status=todo", dbTitle, dbStatus)
	}

	// 5. タスクを完了する。
	rec, _ = doRequest(t, router, http.MethodPatch, "/api/v1/users/"+userID+"/tasks/"+createdTask.ID+"/complete", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH complete status: got %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body)
	}

	if err := db.QueryRow(`SELECT status FROM tasks WHERE id = $1`, createdTask.ID).Scan(&dbStatus); err != nil {
		t.Fatalf("query task status: %v", err)
	}
	if dbStatus != "done" {
		t.Errorf("task status in DB after complete: got %q, want done", dbStatus)
	}

	// 6. 今日の統計に完了したタスクが反映されていることを確認する(tasks テーブルを都度集計するクエリの動作確認)。
	rec, env = doRequest(t, router, http.MethodGet, "/api/v1/users/"+userID+"/stats/today", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /stats/today status: got %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body)
	}
	var stats struct {
		Tasks struct {
			CompletedCount int `json:"completedCount"`
			TotalCount     int `json:"totalCount"`
			CompletionRate int `json:"completionRate"`
		} `json:"tasks"`
	}
	json.Unmarshal(env.Data, &stats)
	if stats.Tasks.CompletedCount != 1 || stats.Tasks.TotalCount != 1 || stats.Tasks.CompletionRate != 100 {
		t.Errorf("unexpected today stats: %+v", stats.Tasks)
	}

	// 7. タスク一覧に自分のタスクのみ含まれることを確認する。
	rec, env = doRequest(t, router, http.MethodGet, "/api/v1/users/"+userID+"/tasks", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /tasks status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var tasks []struct {
		ID string `json:"id"`
	}
	json.Unmarshal(env.Data, &tasks)
	if len(tasks) != 1 || tasks[0].ID != createdTask.ID {
		t.Errorf("unexpected task list: %+v", tasks)
	}
}

// TestUserFlow_TasksAreIsolatedPerUser confirms that one user cannot read or
// complete another user's task, verifying authorization against real DB rows
// rather than fakes.
func TestUserFlow_TasksAreIsolatedPerUser(t *testing.T) {
	db := setupTestDB(t)
	router := newRouter(db)

	_, ownerEnv := doRequest(t, router, http.MethodPost, "/api/v1/users",
		`{"name":"オーナー","presidentName":"オーナー社長"}`)
	var owner struct {
		UserID string `json:"userId"`
	}
	json.Unmarshal(ownerEnv.Data, &owner)
	t.Cleanup(func() { cleanupUser(t, db, owner.UserID) })

	_, otherEnv := doRequest(t, router, http.MethodPost, "/api/v1/users",
		`{"name":"他人","presidentName":"他人社長"}`)
	var other struct {
		UserID string `json:"userId"`
	}
	json.Unmarshal(otherEnv.Data, &other)
	t.Cleanup(func() { cleanupUser(t, db, other.UserID) })

	rec, taskEnv := doRequest(t, router, http.MethodPost, "/api/v1/users/"+owner.UserID+"/tasks", `{"title":"オーナー専用タスク"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /tasks status: got %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body)
	}
	var ownerTask struct {
		ID string `json:"id"`
	}
	json.Unmarshal(taskEnv.Data, &ownerTask)

	// 他人がオーナーのタスクを完了しようとすると 403 になる。
	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/"+other.UserID+"/tasks/"+ownerTask.ID+"/complete", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("PATCH complete (other user) status: got %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body)
	}
	if env.Error == nil || env.Error.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN code, got %+v", env.Error)
	}

	// DB 上でもタスクは todo のままであること(他人の操作で状態が変わっていない)を確認する。
	var dbStatus string
	if err := db.QueryRow(`SELECT status FROM tasks WHERE id = $1`, ownerTask.ID).Scan(&dbStatus); err != nil {
		t.Fatalf("query task status: %v", err)
	}
	if dbStatus != "todo" {
		t.Errorf("task status in DB: got %q, want todo (should be unaffected by other user's request)", dbStatus)
	}

	// 他人のタスク一覧にはオーナーのタスクが含まれない。
	rec, env = doRequest(t, router, http.MethodGet, "/api/v1/users/"+other.UserID+"/tasks", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /tasks (other user) status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var otherTasks []map[string]any
	json.Unmarshal(env.Data, &otherTasks)
	if len(otherTasks) != 0 {
		t.Errorf("expected other user's task list to be empty, got %+v", otherTasks)
	}
}

// TestUserFlow_SelectingUnknownCharacterIsRejected confirms that selecting a
// non-existent character is rejected before the DB row is changed.
func TestUserFlow_SelectingUnknownCharacterIsRejected(t *testing.T) {
	db := setupTestDB(t)
	router := newRouter(db)

	_, userEnv := doRequest(t, router, http.MethodPost, "/api/v1/users",
		`{"name":"統合テスト次郎","presidentName":"統合テスト社長2"}`)
	var u struct {
		UserID string `json:"userId"`
	}
	json.Unmarshal(userEnv.Data, &u)
	t.Cleanup(func() { cleanupUser(t, db, u.UserID) })

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/"+u.UserID+"/selected-character",
		`{"characterId":"char_does_not_exist"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PATCH selected-character status: got %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body)
	}
	if env.Error == nil || env.Error.Code != "CHARACTER_NOT_FOUND" {
		t.Errorf("expected CHARACTER_NOT_FOUND code, got %+v", env.Error)
	}

	var selectedCharID sql.NullString
	if err := db.QueryRow(`SELECT selected_character_id FROM users WHERE id = $1`, u.UserID).Scan(&selectedCharID); err != nil {
		t.Fatalf("query selected_character_id: %v", err)
	}
	if selectedCharID.Valid {
		t.Errorf("selected_character_id in DB should remain unset, got %v", selectedCharID)
	}
}
