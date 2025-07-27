package withdrawal

import (
	"errors"
	"time"

	"github.com/radiophysiker/d56/internal/domain/order"
	"github.com/radiophysiker/d56/internal/domain/user"
)

type WithdrawalID int64

type Withdrawal struct {
	id          WithdrawalID
	userID      user.UserID
	orderNumber string
	amount      float64
	processedAt time.Time
}

func NewWithdrawal(userID user.UserID, orderNumber string, amount float64) (*Withdrawal, error) {
	if userID.String() == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if orderNumber == "" {
		return nil, errors.New("order number cannot be empty")
	}

	if !order.IsValidOrderNumber(orderNumber) {
		return nil, errors.New("invalid order number format")
	}

	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	return &Withdrawal{
		userID:      userID,
		orderNumber: orderNumber,
		amount:      amount,
		processedAt: time.Now(),
	}, nil
}

func (w *Withdrawal) ID() WithdrawalID {
	return w.id
}

func (w *Withdrawal) UserID() user.UserID {
	return w.userID
}

func (w *Withdrawal) OrderNumber() string {
	return w.orderNumber
}

func (w *Withdrawal) Amount() float64 {
	return w.amount
}

func (w *Withdrawal) ProcessedAt() time.Time {
	return w.processedAt
}

// NewWithdrawalFromRepository создает списание из данных репозитория
func NewWithdrawalFromRepository(id WithdrawalID, userID user.UserID, orderNumber string, amount float64, processedAt time.Time) *Withdrawal {
	return &Withdrawal{
		id:          id,
		userID:      userID,
		orderNumber: orderNumber,
		amount:      amount,
		processedAt: processedAt,
	}
}
