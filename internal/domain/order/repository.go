package order

import (
	"context"

	"github.com/radiophysiker/d56/internal/domain/user"
)

type Repository interface {
	Save(ctx context.Context, order *Order) error
	FindByNumber(ctx context.Context, number string) (*Order, error)
	FindByUserID(ctx context.Context, userID user.UserID) ([]*Order, error)
	FindPendingOrders(ctx context.Context) ([]*Order, error)
	Update(ctx context.Context, order *Order) error
}
