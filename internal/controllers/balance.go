package controllers

import (
	"context"
	"net/http"

	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/server"
	"github.com/gopybara/httpbara"
	"github.com/gopybara/httpbara/casual"
	"go.uber.org/fx"
)

type balanceControllerDescription struct {
	BalanceAPI  httpbara.Group `group:"/api/user"`
	GetBalance  httpbara.Route `route:"GET /balance" group:"balanceapi" middlewares:"jwt"`
	Withdraw    httpbara.Route `route:"POST /balance/withdraw" group:"balanceapi" middlewares:"jwt"`
	Withdrawals httpbara.Route `route:"GET /withdrawals" group:"balanceapi" middlewares:"jwt"`
}

type newBalanceControllerIn struct {
	fx.In

	Logger         httpbara.Logger
	BalanceService domain.BalanceServiceI
}

type balanceController struct {
	balanceControllerDescription

	logger         httpbara.Logger
	balanceService domain.BalanceServiceI
}

type WithdrawalsResponse struct {
	Withdrawals []models.Withdrawal
}

func (wr *WithdrawalsResponse) StatusCode() int {
	if len(wr.Withdrawals) == 0 {
		return http.StatusNoContent
	}
	return http.StatusOK
}

type WithdrawResponse struct{}

func (wr *WithdrawResponse) StatusCode() int {
	return http.StatusOK
}

func NewBalanceController(in newBalanceControllerIn) (server.AsHandlerOut, error) {
	return server.AsHandler(&balanceController{
		logger:         in.Logger,
		balanceService: in.BalanceService,
	})
}

func (bc *balanceController) GetBalance(ctx context.Context, _ *struct{}) (*models.Balance, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		bc.logger.Warn("User not authenticated")
		return nil, casual.NewHTTPErrorFromMessage(401, "Unauthorized")
	}

	balance, err := bc.balanceService.GetBalance(userID)
	if err != nil {
		bc.logger.Error("Failed to get balance", "error", err, "user_id", userID)
		return nil, casual.NewHTTPErrorFromMessage(500, "Failed to get balance")
	}

	bc.logger.Info("Balance retrieved successfully", "user_id", userID, "balance", balance.Current)
	return balance, nil
}

func (bc *balanceController) Withdraw(ctx context.Context, req *models.WithdrawalRequest) (*WithdrawResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		bc.logger.Warn("User not authenticated")
		return nil, casual.NewHTTPErrorFromMessage(401, "Unauthorized")
	}

	err := bc.balanceService.WithdrawBalance(userID, req.Order, req.Sum)
	if err != nil {
		bc.logger.Error("Failed to withdraw", "error", err, "user_id", userID, "order", req.Order, "amount", req.Sum)
		return nil, casual.NewHTTPErrorFromMessage(500, "Failed to withdraw")
	}

	bc.logger.Info("Withdrawal successful", "user_id", userID, "order", req.Order, "amount", req.Sum)
	return &WithdrawResponse{}, nil
}

func (bc *balanceController) Withdrawals(ctx context.Context, _ *struct{}) (*WithdrawalsResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		bc.logger.Warn("User not authenticated")
		return nil, casual.NewHTTPErrorFromMessage(401, "Unauthorized")
	}

	withdrawals, err := bc.balanceService.GetWithdrawals(userID)
	if err != nil {
		bc.logger.Error("Failed to get withdrawals", "error", err, "user_id", userID)
		return nil, casual.NewHTTPErrorFromMessage(500, "Failed to get withdrawals")
	}

	bc.logger.Info("Withdrawals retrieved successfully", "count", len(withdrawals), "user_id", userID)
	return &WithdrawalsResponse{Withdrawals: withdrawals}, nil
}
