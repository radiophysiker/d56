package service

import (
	"context"

	"github.com/radiophysiker/d56/internal/domain/user"
	"github.com/radiophysiker/d56/internal/domain/withdrawal"
)

type WithdrawalService struct {
	withdrawalRepo withdrawal.Repository
	userRepo       user.Repository
}

func NewWithdrawalService(withdrawalRepo withdrawal.Repository, userRepo user.Repository) *WithdrawalService {
	return &WithdrawalService{
		withdrawalRepo: withdrawalRepo,
		userRepo:       userRepo,
	}
}

func (s *WithdrawalService) WithdrawBalance(ctx context.Context, userID user.UserID, orderNumber string, amount float64) (*withdrawal.Withdrawal, error) {
	// Создаем объект списания
	w, err := withdrawal.NewWithdrawal(userID, orderNumber, amount)
	if err != nil {
		return nil, err
	}

	// Получаем пользователя и проверяем баланс
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Списываем баланс
	if err := user.WithdrawBalance(amount); err != nil {
		return nil, err
	}

	// Сохраняем изменения пользователя
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	// Сохраняем запись о списании
	if err := s.withdrawalRepo.Save(ctx, w); err != nil {
		return nil, err
	}

	return w, nil
}

func (s *WithdrawalService) GetUserWithdrawals(ctx context.Context, userID user.UserID) ([]*withdrawal.Withdrawal, error) {
	return s.withdrawalRepo.FindByUserID(ctx, userID)
}
