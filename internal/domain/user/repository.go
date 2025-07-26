package user

import "context"

type Repository interface {
	Save(ctx context.Context, user *User) error
	FindByLogin(ctx context.Context, login string) (*User, error)
	FindByID(ctx context.Context, id UserID) (*User, error)
}
