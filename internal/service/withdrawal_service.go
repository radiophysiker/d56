package service

import (
	"context"

	"github.com/radiophysiker/d56/internal/domain/user"
	"github.com/radiophysiker/d56/internal/domain/withdrawal"
	"github.com/radiophysiker/d56/internal/infrastructure/database"
)

type WithdrawalService struct {
	withdrawalRepo withdrawal.Repository
	userRepo       user.Repository
	txManager      database.TransactionManager
}

func NewWithdrawalService(
	withdrawalRepo withdrawal.Repository,
	userRepo user.Repository,
	txManager database.TransactionManager,
) *WithdrawalService {
	return &WithdrawalService{
		withdrawalRepo: withdrawalRepo,
		userRepo:       userRepo,
		txManager:      txManager,
	}
}

func (s *WithdrawalService) WithdrawBalance(ctx context.Context, userID user.UserID, orderNumber string, amount float64) (*withdrawal.Withdrawal, error) {
	// Создаем объект списания для валидации
	w, err := withdrawal.NewWithdrawal(userID, orderNumber, amount)
	if err != nil {
		return nil, err
	}

	// Выполняем операцию в рамках транзакции согласно DDD
	var resultWithdrawal *withdrawal.Withdrawal
	err = s.txManager.WithTransactionContext(ctx, func(txCtx context.Context) error {
		// Получаем пользователя
		user, err := s.userRepo.FindByID(txCtx, userID)
		if err != nil {
			return err
		}

		// Списываем баланс (Domain Logic)
		if err := user.WithdrawBalance(amount); err != nil {
			return err
		}

		// Сохраняем изменения пользователя (автоматически в транзакции через контекст)
		if err := s.userRepo.Save(txCtx, user); err != nil {
			return err
		}

		// Сохраняем запись о списании (автоматически в транзакции через контекст)
		if err := s.withdrawalRepo.Save(txCtx, w); err != nil {
			return err
		}

		resultWithdrawal = w
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resultWithdrawal, nil
}

func (s *WithdrawalService) GetUserWithdrawals(ctx context.Context, userID user.UserID) ([]*withdrawal.Withdrawal, error) {
	return s.withdrawalRepo.FindByUserID(ctx, userID)
}
