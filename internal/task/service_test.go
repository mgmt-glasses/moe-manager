package task_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mgmt-glasses/moe-manager/internal/task"
	"github.com/mgmt-glasses/moe-manager/internal/task/testutil"
)

func TestService_Create_OK(t *testing.T) {
	repo := testutil.NewFakeRepository()
	svc := task.NewService(repo)

	got, err := svc.Create(context.Background(), task.CreateInput{UserID: "u_001", Title: "資料作成"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID == "" {
		t.Error("expected generated ID")
	}
	if got.Status != task.StatusTodo {
		t.Errorf("status: got %q, want %q", got.Status, task.StatusTodo)
	}
	if _, err := repo.FindByID(context.Background(), got.ID); err != nil {
		t.Errorf("expected task to be persisted: %v", err)
	}
}

func TestService_Create_MissingTitle(t *testing.T) {
	svc := task.NewService(testutil.NewFakeRepository())

	_, err := svc.Create(context.Background(), task.CreateInput{UserID: "u_001"})
	if !errors.Is(err, task.ErrBadInput) {
		t.Errorf("expected ErrBadInput, got %v", err)
	}
}

func TestService_List_OnlyReturnsOwnTasks(t *testing.T) {
	repo := testutil.NewFakeRepository(
		task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusTodo},
		task.Task{ID: "t_002", UserID: "u_002", Title: "B", Status: task.StatusTodo},
	)
	svc := task.NewService(repo)

	got, err := svc.List(context.Background(), "u_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "t_001" {
		t.Errorf("expected only u_001's task, got %+v", got)
	}
}

func TestService_Complete_OK(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusTodo})
	svc := task.NewService(repo)

	got, err := svc.Complete(context.Background(), "u_001", "t_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != task.StatusDone {
		t.Errorf("status: got %q, want %q", got.Status, task.StatusDone)
	}
	if got.CompletedAt == nil {
		t.Error("expected completedAt to be set")
	}
}

func TestService_Complete_NotFound(t *testing.T) {
	svc := task.NewService(testutil.NewFakeRepository())

	_, err := svc.Complete(context.Background(), "u_001", "t_missing")
	if !errors.Is(err, task.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Complete_ForbiddenForOtherUser(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusTodo})
	svc := task.NewService(repo)

	_, err := svc.Complete(context.Background(), "u_002", "t_001")
	if !errors.Is(err, task.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Reopen_OK(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusDone})
	svc := task.NewService(repo)

	got, err := svc.Reopen(context.Background(), "u_001", "t_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != task.StatusTodo {
		t.Errorf("status: got %q, want %q", got.Status, task.StatusTodo)
	}
	if got.CompletedAt != nil {
		t.Error("expected completedAt to be cleared")
	}
}

func TestService_Reopen_ForbiddenForOtherUser(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusDone})
	svc := task.NewService(repo)

	_, err := svc.Reopen(context.Background(), "u_002", "t_001")
	if !errors.Is(err, task.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Delete_OK(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusTodo})
	svc := task.NewService(repo)

	if err := svc.Delete(context.Background(), "u_001", "t_001"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := repo.FindByID(context.Background(), "t_001"); !errors.Is(err, task.ErrNotFound) {
		t.Errorf("expected task to be removed, got err=%v", err)
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	svc := task.NewService(testutil.NewFakeRepository())

	err := svc.Delete(context.Background(), "u_001", "t_missing")
	if !errors.Is(err, task.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Delete_ForbiddenForOtherUser(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusTodo})
	svc := task.NewService(repo)

	err := svc.Delete(context.Background(), "u_002", "t_001")
	if !errors.Is(err, task.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
