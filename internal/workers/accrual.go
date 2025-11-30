package workers

import (
	"context"
	"time"

	"github.com/SZabrodskii/gophermart-stas/internal/accrual"
	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

type AccrualWorkerI interface {
	Start(ctx context.Context) error
	Stop()
}

type newAccrualWorkerIn struct {
	fx.In

	Logger         httpbara.Logger
	OrderService   domain.OrderServiceI
	BalanceService domain.BalanceServiceI
	AccrualClient  accrual.ClientI
}

type accrualWorker struct {
	logger         httpbara.Logger
	orderService   domain.OrderServiceI
	balanceService domain.BalanceServiceI
	accrualClient  accrual.ClientI
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
		aw.logger.Error("Failed to get orders for processing", "error", err)
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
		aw.logger.Error("Failed to get order info from accrual", "error", err, "order", order.Number)
		return
	}

	// Map external accrual statuses to internal order statuses
	internalStatus := aw.mapExternalStatus(accrualInfo.Status)

	accrualEqual := (order.Accrual == nil && accrualInfo.Accrual == nil) ||
		(order.Accrual != nil && accrualInfo.Accrual != nil && *order.Accrual == *accrualInfo.Accrual)

	if internalStatus == order.Status && accrualEqual {
		return
	}

	aw.logger.Info("Updating order", "order", order.Number, "old_status", order.Status, "new_status", internalStatus, "external_status", accrualInfo.Status)

	var accrualValue float64
	if accrualInfo.Accrual != nil {
		accrualValue = *accrualInfo.Accrual
	}

	err = aw.orderService.UpdateOrderStatus(order.Number, internalStatus, accrualValue)
	if err != nil {
		aw.logger.Error("Failed to update order", "error", err, "order", order.Number)
		return
	}

	if internalStatus == models.OrderStatusProcessed && accrualInfo.Accrual != nil && *accrualInfo.Accrual > 0 {
		err = aw.balanceService.AddBalance(order.UserID, *accrualInfo.Accrual)
		if err != nil {
			aw.logger.Error("Failed to add balance", "error", err, "user_id", order.UserID, "amount", *accrualInfo.Accrual)
			return
		}
		aw.logger.Info("Balance added", "user_id", order.UserID, "amount", *accrualInfo.Accrual, "order", order.Number)
	}
}

// mapExternalStatus maps external accrual system statuses to internal order statuses
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
		// Log warning for unknown status and return as-is
		aw.logger.Warn("Unknown external accrual status", "status", externalStatus)
		return externalStatus
	}
}
