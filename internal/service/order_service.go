package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/radiophysiker/d56/internal/domain/order"
	"github.com/radiophysiker/d56/internal/domain/user"
)

type OrderService struct {
	orderRepo order.Repository
	userRepo  user.Repository
}

func NewOrderService(orderRepo order.Repository, userRepo user.Repository) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		userRepo:  userRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID user.UserID, number string) (*order.Order, bool, error) {
	// Проверяем, не существует ли заказ с таким номером
	existingOrder, err := s.orderRepo.FindByNumber(ctx, number)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	if existingOrder != nil {
		// Если заказ уже существует и принадлежит тому же пользователю
		if existingOrder.UserID() == userID {
			return existingOrder, false, nil // Возвращаем существующий заказ, не новый
		}
		// Если заказ принадлежит другому пользователю
		return nil, false, errors.New("order already uploaded by another user")
	}

	// Создаем новый заказ
	newOrder, err := order.NewOrder(userID, number)
	if err != nil {
		return nil, false, err
	}

	if err := s.orderRepo.Save(ctx, newOrder); err != nil {
		return nil, false, err
	}

	return newOrder, true, nil // Возвращаем новый заказ
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID user.UserID) ([]*order.Order, error) {
	return s.orderRepo.FindByUserID(ctx, userID)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderNumber string, status order.Status, accrual *float64) error {
	orderEntity, err := s.orderRepo.FindByNumber(ctx, orderNumber)
	if err != nil {
		return err
	}

	orderEntity.SetStatus(status)
	if accrual != nil && status == order.StatusProcessed {
		orderEntity.SetAccrual(*accrual)

		// Если заказ обработан и есть начисление, обновляем баланс пользователя
		if *accrual > 0 {
			user, err := s.userRepo.FindByID(ctx, orderEntity.UserID())
			if err != nil {
				return err
			}

			user.AddBalance(*accrual)
			if err := s.userRepo.Save(ctx, user); err != nil {
				return err
			}
		}
	}

	return s.orderRepo.Update(ctx, orderEntity)
}

func (s *OrderService) GetPendingOrders(ctx context.Context) ([]*order.Order, error) {
	return s.orderRepo.FindPendingOrders(ctx)
}
