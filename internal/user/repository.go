package user

import "context"

type Repository interface {
	Create(ctx context.Context, u User) error
	FindByID(ctx context.Context, id string) (User, error)
	Update(ctx context.Context, u User) error
}

// CharacterValidator は user パッケージが character モジュールに直接依存しないための interface。
type CharacterValidator interface {
	Exists(ctx context.Context, characterID string) (bool, error)
}
