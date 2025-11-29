package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/server"
	"github.com/SZabrodskii/gophermart-stas/internal/services"
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

type WithdrawalsResponse []models.Withdrawal

func (wr *WithdrawalsResponse) StatusCode() int {
	if len(*wr) == 0 {
		return http.StatusNoContent
	}
	return http.StatusOK
}

type WithdrawResponse struct {
	Code int
}

func (wr *WithdrawResponse) StatusCode() int {
	return wr.Code
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
		return &WithdrawResponse{Code: http.StatusUnauthorized}, nil
	}

	err := bc.balanceService.WithdrawBalance(userID, req.Order, req.Sum)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientFunds):
			bc.logger.Warn("Insufficient funds for withdrawal", "user_id", userID, "amount", req.Sum)
			return &WithdrawResponse{Code: http.StatusPaymentRequired}, nil
		case errors.Is(err, services.ErrInvalidOrderNumber):
			bc.logger.Warn("Invalid order number format", "order", req.Order)
			return &WithdrawResponse{Code: http.StatusUnprocessableEntity}, nil
		default:
			bc.logger.Error("Failed to withdraw", "error", err, "user_id", userID, "order", req.Order, "amount", req.Sum)
			return &WithdrawResponse{Code: http.StatusInternalServerError}, nil
		}
	}

	bc.logger.Info("Withdrawal successful", "user_id", userID, "order", req.Order, "amount", req.Sum)
	return &WithdrawResponse{Code: http.StatusOK}, nil
}

func (bc *balanceController) Withdrawals(ctx context.Context, _ *struct{}) (WithdrawalsResponse, error) {
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
	return WithdrawalsResponse(withdrawals), nil
}
