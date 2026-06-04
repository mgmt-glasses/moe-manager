package character

import "context"

type Repository interface {
	List(ctx context.Context) ([]Character, error)
	FindByID(ctx context.Context, id string) (Character, error)
	Exists(ctx context.Context, id string) (bool, error)
}
