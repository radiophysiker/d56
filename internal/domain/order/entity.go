package order

import (
	"errors"
	"strconv"
	"time"

	"github.com/radiophysiker/d56/internal/domain/user"
)

type OrderID int64

type Status string

const (
	StatusNew        Status = "NEW"
	StatusProcessing Status = "PROCESSING"
	StatusInvalid    Status = "INVALID"
	StatusProcessed  Status = "PROCESSED"
)

type Order struct {
	id          OrderID
	userID      user.UserID
	number      string
	status      Status
	accrual     *float64
	uploadedAt  time.Time
	processedAt *time.Time
}

func NewOrder(userID user.UserID, number string) (*Order, error) {
	if userID.String() == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if number == "" {
		return nil, errors.New("order number cannot be empty")
	}

	if !IsValidOrderNumber(number) {
		return nil, errors.New("invalid order number format")
	}

	return &Order{
		userID:     userID,
		number:     number,
		status:     StatusNew,
		uploadedAt: time.Now(),
	}, nil
}

func (o *Order) ID() OrderID {
	return o.id
}

func (o *Order) UserID() user.UserID {
	return o.userID
}

func (o *Order) Number() string {
	return o.number
}

func (o *Order) Status() Status {
	return o.status
}

func (o *Order) Accrual() *float64 {
	return o.accrual
}

func (o *Order) UploadedAt() time.Time {
	return o.uploadedAt
}

func (o *Order) ProcessedAt() *time.Time {
	return o.processedAt
}

func (o *Order) SetStatus(status Status) {
	o.status = status
	if status == StatusProcessed || status == StatusInvalid {
		now := time.Now()
		o.processedAt = &now
	}
}

func (o *Order) SetAccrual(accrual float64) {
	o.accrual = &accrual
}

// IsValidOrderNumber проверяет номер заказа по алгоритму Луна
func IsValidOrderNumber(number string) bool {
	if len(number) == 0 {
		return false
	}

	// Проверяем, что все символы - цифры
	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Алгоритм Луна
	sum := 0
	alternate := false

	// Проходим цифры справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// NewOrderFromRepository создает заказ из данных репозитория
func NewOrderFromRepository(id OrderID, userID user.UserID, number string, status Status, accrual *float64, uploadedAt time.Time, processedAt *time.Time) *Order {
	return &Order{
		id:          id,
		userID:      userID,
		number:      number,
		status:      status,
		accrual:     accrual,
		uploadedAt:  uploadedAt,
		processedAt: processedAt,
	}
}
