package withdrawal

import (
	"context"

	"github.com/radiophysiker/d56/internal/domain/user"
)

type Repository interface {
	Save(ctx context.Context, withdrawal *Withdrawal) error
	FindByUserID(ctx context.Context, userID user.UserID) ([]*Withdrawal, error)
}
