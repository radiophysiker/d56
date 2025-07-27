package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/radiophysiker/d56/internal/domain/order"
	"github.com/radiophysiker/d56/internal/domain/user"
	"github.com/radiophysiker/d56/internal/infrastructure/database"
)

type OrderService struct {
	orderRepo order.Repository
	userRepo  user.Repository
	txManager database.TransactionManager
}

func NewOrderService(
	orderRepo order.Repository,
	userRepo user.Repository,
	txManager database.TransactionManager,
) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		userRepo:  userRepo,
		txManager: txManager,
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

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderNumber string, status order.Status, accrual *float64) error {
	// Выполняем операцию в рамках транзакции согласно DDD
	return s.txManager.WithTransactionContext(ctx, func(txCtx context.Context) error {
		// Получаем заказ
		orderEntity, err := s.orderRepo.FindByNumber(txCtx, orderNumber)
		if err != nil {
			return err
		}

		// Обновляем статус заказа (Domain Logic)
		orderEntity.SetStatus(status)
		if accrual != nil && status == order.StatusProcessed {
			orderEntity.SetAccrual(*accrual)

			// Если заказ обработан и есть начисление, обновляем баланс пользователя
			if *accrual > 0 {
				user, err := s.userRepo.FindByID(txCtx, orderEntity.UserID())
				if err != nil {
					return err
				}

				// Добавляем баланс (Domain Logic)
				user.AddBalance(*accrual)

				// Сохраняем пользователя (автоматически в транзакции через контекст)
				if err := s.userRepo.Save(txCtx, user); err != nil {
					return err
				}
			}
		}

		// Сохраняем заказ (автоматически в транзакции через контекст)
		return s.orderRepo.Update(txCtx, orderEntity)
	})
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID user.UserID) ([]*order.Order, error) {
	return s.orderRepo.FindByUserID(ctx, userID)
}

func (s *OrderService) GetPendingOrders(ctx context.Context) ([]*order.Order, error) {
	return s.orderRepo.FindPendingOrders(ctx)
}
