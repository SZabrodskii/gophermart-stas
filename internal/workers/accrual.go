package workers

import (
	"context"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/accrual"
	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type AccrualWorkerI interface {
	Start(ctx context.Context) error
	Stop()
}

type newAccrualWorkerIn struct {
	fx.In

	Logger         *zap.Logger
	OrderService   domain.OrderServiceI
	BalanceService domain.BalanceServiceI
	AccrualClient  accrual.Client
}

type accrualWorker struct {
	logger         *zap.Logger
	orderService   domain.OrderServiceI
	balanceService domain.BalanceServiceI
	accrualClient  accrual.Client
	stopCh         chan struct{}
}

func NewAccrualWorker(in newAccrualWorkerIn) AccrualWorkerI {
	return &accrualWorker{
		logger:         in.Logger,
		orderService:   in.OrderService,
		balanceService: in.BalanceService,
		accrualClient:  in.AccrualClient,
		stopCh:         make(chan struct{}),
	}
}

func (aw *accrualWorker) Start(ctx context.Context) error {
	aw.logger.Info("Starting accrual worker")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			aw.logger.Info("Accrual worker stopped by context")
			return ctx.Err()
		case <-aw.stopCh:
			aw.logger.Info("Accrual worker stopped")
			return nil
		case <-ticker.C:
			aw.processOrders(ctx)
		}
	}
}

func (aw *accrualWorker) Stop() {
	close(aw.stopCh)
}

func (aw *accrualWorker) processOrders(ctx context.Context) {
	orders, err := aw.orderService.GetOrdersForProcessing()
	if err != nil {
		aw.logger.Error("Failed to get orders for processing", zap.Error(err))
		return
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return
		default:
			aw.processOrder(ctx, order)
		}
	}
}

func (aw *accrualWorker) processOrder(ctx context.Context, order *models.Order) {
	accrualInfo, err := aw.accrualClient.GetOrderInfo(ctx, order.Number)
	if err != nil {
		aw.logger.Error("Failed to get order info from accrual", zap.Error(err), zap.String("order", order.Number))
		return
	}

	if accrualInfo == nil {
		return
	}

	internalStatus := aw.mapExternalStatus(accrualInfo.Status)

	accrualEqual := (order.Accrual == nil && accrualInfo.Accrual == nil) ||
		(order.Accrual != nil && accrualInfo.Accrual != nil && *order.Accrual == *accrualInfo.Accrual)

	if internalStatus == order.Status && accrualEqual {
		return
	}

	aw.logger.Info("Updating order",
		zap.String("order", order.Number),
		zap.String("old_status", order.Status),
		zap.String("new_status", internalStatus),
		zap.String("external_status", accrualInfo.Status))

	var accrualValue float64
	if accrualInfo.Accrual != nil {
		accrualValue = *accrualInfo.Accrual
	}

	err = aw.orderService.UpdateOrderStatus(order.Number, internalStatus, accrualValue)
	if err != nil {
		aw.logger.Error("Failed to update order", zap.Error(err), zap.String("order", order.Number))
		return
	}

	if internalStatus == models.OrderStatusProcessed && accrualInfo.Accrual != nil && *accrualInfo.Accrual > 0 {
		err = aw.balanceService.AddBalance(order.UserID, *accrualInfo.Accrual)
		if err != nil {
			aw.logger.Error("Failed to add balance", zap.Error(err), zap.Uint("user_id", order.UserID), zap.Float64("amount", *accrualInfo.Accrual))
			return
		}
		aw.logger.Info("Balance added", zap.Uint("user_id", order.UserID), zap.Float64("amount", *accrualInfo.Accrual), zap.String("order", order.Number))
	}
}

func (aw *accrualWorker) mapExternalStatus(externalStatus string) string {
	switch externalStatus {
	case accrual.StatusRegistered:
		return models.OrderStatusNew
	case accrual.StatusInvalid:
		return models.OrderStatusInvalid
	case accrual.StatusProcessing:
		return models.OrderStatusProcessing
	case accrual.StatusProcessed:
		return models.OrderStatusProcessed
	default:
		aw.logger.Warn("Unknown external accrual status", zap.String("status", externalStatus))
		return externalStatus
	}
}
