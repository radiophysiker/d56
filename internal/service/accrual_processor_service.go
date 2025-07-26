package service

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/radiophysiker/d56/internal/infrastructure/accrual"
)

type AccrualProcessorService struct {
	accrualClient *accrual.Client
	orderService  *OrderService
	logger        *zap.Logger
}

func NewAccrualProcessorService(accrualClient *accrual.Client, orderService *OrderService, logger *zap.Logger) *AccrualProcessorService {
	return &AccrualProcessorService{
		accrualClient: accrualClient,
		orderService:  orderService,
		logger:        logger,
	}
}

func (s *AccrualProcessorService) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Проверяем каждые 10 секунд
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Accrual processor service stopped")
			return
		case <-ticker.C:
			s.processOrders(ctx)
		}
	}
}

func (s *AccrualProcessorService) processOrders(ctx context.Context) {
	orders, err := s.orderService.GetPendingOrders(ctx)
	if err != nil {
		s.logger.Error("Failed to get pending orders", zap.Error(err))
		return
	}

	for _, order := range orders {
		s.processOrder(ctx, order.Number())
	}
}

func (s *AccrualProcessorService) processOrder(ctx context.Context, orderNumber string) {
	accrualResp, err := s.accrualClient.GetOrderInfo(ctx, orderNumber)
	if err != nil {
		if strings.Contains(err.Error(), "rate limit exceeded") {
			s.logger.Warn("Rate limit exceeded, will retry later", zap.String("order", orderNumber))
			return
		}
		s.logger.Error("Failed to get order info from accrual system",
			zap.String("order", orderNumber), zap.Error(err))
		return
	}

	if accrualResp == nil {
		// Заказ не найден в системе расчета, оставляем как NEW
		return
	}

	status := s.accrualClient.ConvertToOrderStatus(accrualResp.Status)

	err = s.orderService.UpdateOrderStatus(ctx, orderNumber, status, accrualResp.Accrual)
	if err != nil {
		s.logger.Error("Failed to update order status",
			zap.String("order", orderNumber), zap.Error(err))
		return
	}

	s.logger.Info("Updated order status",
		zap.String("order", orderNumber),
		zap.String("status", string(status)))
}
