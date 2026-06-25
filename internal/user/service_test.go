package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mgmt-glasses/moe-manager/internal/user"
	"github.com/mgmt-glasses/moe-manager/internal/user/testutil"
)

func TestService_Create_OK(t *testing.T) {
	repo := testutil.NewFakeRepository()
	charValid := testutil.NewFakeCharacterValidator("char_001")

	svc := user.NewService(repo, charValid)
	charID := "char_001"
	got, err := svc.Create(context.Background(), user.CreateInput{
		Name:                       "山田太郎",
		PresidentName:              "山田社長",
		TargetEntertainmentMinutes: 90,
		SelectedCharacterID:        &charID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID == "" {
		t.Error("expected generated ID")
	}
	if got.TargetEntertainmentMinutes != 90 {
		t.Errorf("targetEntertainmentMinutes: got %d, want 90", got.TargetEntertainmentMinutes)
	}
	if got.SelectedCharacterID == nil || *got.SelectedCharacterID != charID {
		t.Errorf("selectedCharacterId: got %v, want %q", got.SelectedCharacterID, charID)
	}
	if _, err := repo.FindByID(context.Background(), got.ID); err != nil {
		t.Errorf("expected user to be persisted: %v", err)
	}
}

func TestService_Create_UsesProvidedUserID(t *testing.T) {
	repo := testutil.NewFakeRepository()
	svc := user.NewService(repo, testutil.NewFakeCharacterValidator())

	got, err := svc.Create(context.Background(), user.CreateInput{
		UserID:        "firebase_uid_001",
		Name:          "山田太郎",
		PresidentName: "山田社長",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "firebase_uid_001" {
		t.Errorf("id: got %q, want firebase_uid_001", got.ID)
	}
	if _, err := repo.FindByID(context.Background(), "firebase_uid_001"); err != nil {
		t.Errorf("expected user to be persisted by provided id: %v", err)
	}
}

func TestService_Create_DefaultsTargetEntertainmentMinutes(t *testing.T) {
	svc := user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())

	got, err := svc.Create(context.Background(), user.CreateInput{Name: "山田太郎", PresidentName: "山田社長"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TargetEntertainmentMinutes != 120 {
		t.Errorf("default targetEntertainmentMinutes: got %d, want 120", got.TargetEntertainmentMinutes)
	}
}

func TestService_Create_MissingName(t *testing.T) {
	svc := user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())

	_, err := svc.Create(context.Background(), user.CreateInput{PresidentName: "山田社長"})
	if !errors.Is(err, user.ErrBadInput) {
		t.Errorf("expected ErrBadInput, got %v", err)
	}
}

func TestService_Create_MissingPresidentName(t *testing.T) {
	svc := user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())

	_, err := svc.Create(context.Background(), user.CreateInput{Name: "山田太郎"})
	if !errors.Is(err, user.ErrBadInput) {
		t.Errorf("expected ErrBadInput, got %v", err)
	}
}

func TestService_Create_CharacterNotFound(t *testing.T) {
	svc := user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())

	charID := "char_missing"
	_, err := svc.Create(context.Background(), user.CreateInput{
		Name: "山田太郎", PresidentName: "山田社長", SelectedCharacterID: &charID,
	})
	if !errors.Is(err, user.ErrCharacterNotFound) {
		t.Errorf("expected ErrCharacterNotFound, got %v", err)
	}
}

func TestService_Create_CharacterValidatorError(t *testing.T) {
	charValid := testutil.NewFakeCharacterValidator()
	charValid.Err = errors.New("firestore unavailable")
	svc := user.NewService(testutil.NewFakeRepository(), charValid)

	charID := "char_001"
	_, err := svc.Create(context.Background(), user.CreateInput{
		Name: "山田太郎", PresidentName: "山田社長", SelectedCharacterID: &charID,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_FindByID_OK(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "山田太郎"}
	svc := user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())

	got, err := svc.FindByID(context.Background(), "u_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "山田太郎" {
		t.Errorf("name: got %q, want %q", got.Name, "山田太郎")
	}
}

func TestService_FindByID_NotFound(t *testing.T) {
	svc := user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())

	_, err := svc.FindByID(context.Background(), "u_missing")
	if !errors.Is(err, user.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Update_PartialFields(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "旧名前", PresidentName: "旧社長", TargetEntertainmentMinutes: 60}
	svc := user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())

	newName := "新名前"
	got, err := svc.Update(context.Background(), "u_001", user.UpdateInput{Name: &newName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "新名前" {
		t.Errorf("name: got %q, want %q", got.Name, "新名前")
	}
	if got.PresidentName != "旧社長" {
		t.Errorf("presidentName should be unchanged: got %q", got.PresidentName)
	}
	if got.TargetEntertainmentMinutes != 60 {
		t.Errorf("targetEntertainmentMinutes should be unchanged: got %d", got.TargetEntertainmentMinutes)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	svc := user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())

	newName := "新名前"
	_, err := svc.Update(context.Background(), "u_missing", user.UpdateInput{Name: &newName})
	if !errors.Is(err, user.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Update_EmptyNameRejected(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "旧名前", PresidentName: "旧社長"}
	svc := user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())

	empty := ""
	_, err := svc.Update(context.Background(), "u_001", user.UpdateInput{Name: &empty})
	if !errors.Is(err, user.ErrBadInput) {
		t.Errorf("expected ErrBadInput, got %v", err)
	}
}

func TestService_Update_NonPositiveTargetRejected(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "旧名前", PresidentName: "旧社長"}
	svc := user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())

	zero := 0
	_, err := svc.Update(context.Background(), "u_001", user.UpdateInput{TargetEntertainmentMinutes: &zero})
	if !errors.Is(err, user.ErrBadInput) {
		t.Errorf("expected ErrBadInput, got %v", err)
	}
}

func TestService_Update_CharacterNotFound(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "旧名前", PresidentName: "旧社長"}
	svc := user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())

	charID := "char_missing"
	_, err := svc.Update(context.Background(), "u_001", user.UpdateInput{SelectedCharacterID: &charID})
	if !errors.Is(err, user.ErrCharacterNotFound) {
		t.Errorf("expected ErrCharacterNotFound, got %v", err)
	}
}

func TestService_UpdateSelectedCharacter_OK(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "山田太郎"}
	svc := user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator("char_001"))

	got, err := svc.UpdateSelectedCharacter(context.Background(), "u_001", "char_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SelectedCharacterID == nil || *got.SelectedCharacterID != "char_001" {
		t.Errorf("selectedCharacterId: got %v, want char_001", got.SelectedCharacterID)
	}
}

func TestService_UpdateSelectedCharacter_UserNotFound(t *testing.T) {
	svc := user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator("char_001"))

	_, err := svc.UpdateSelectedCharacter(context.Background(), "u_missing", "char_001")
	if !errors.Is(err, user.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_UpdateSelectedCharacter_CharacterNotFound(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "山田太郎"}
	svc := user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())

	_, err := svc.UpdateSelectedCharacter(context.Background(), "u_001", "char_missing")
	if !errors.Is(err, user.ErrCharacterNotFound) {
		t.Errorf("expected ErrCharacterNotFound, got %v", err)
	}
}
