package character_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mgmt-glasses/moe-manager/internal/character"
	"github.com/mgmt-glasses/moe-manager/internal/character/testutil"
)

var sampleCharacters = []character.Character{
	{ID: "char_001", Name: "さくら", MBTIType: "ISTJ", PersonalityDesc: "真面目で責任感が強い", SpeechStyle: "丁寧で落ち着いた口調"},
	{ID: "char_002", Name: "つむぎ", MBTIType: "ENFP", PersonalityDesc: "明るく社交的", SpeechStyle: "フランクで元気"},
}

func TestService_List_OK(t *testing.T) {
	repo := &testutil.FakeRepository{Characters: sampleCharacters}
	svc := character.NewService(repo)

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(sampleCharacters) {
		t.Fatalf("expected %d characters, got %d", len(sampleCharacters), len(got))
	}
}

func TestService_List_RepositoryError(t *testing.T) {
	repo := &testutil.FakeRepository{ListErr: errors.New("db down")}
	svc := character.NewService(repo)

	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_FindByID_OK(t *testing.T) {
	repo := &testutil.FakeRepository{Characters: sampleCharacters}
	svc := character.NewService(repo)

	got, err := svc.FindByID(context.Background(), "char_002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "つむぎ" {
		t.Errorf("name: got %q, want %q", got.Name, "つむぎ")
	}
}

func TestService_FindByID_NotFound(t *testing.T) {
	repo := &testutil.FakeRepository{Characters: sampleCharacters}
	svc := character.NewService(repo)

	_, err := svc.FindByID(context.Background(), "char_999")
	if !errors.Is(err, character.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_FindByID_RepositoryError(t *testing.T) {
	repo := &testutil.FakeRepository{FindErr: errors.New("db down")}
	svc := character.NewService(repo)

	_, err := svc.FindByID(context.Background(), "char_001")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
