package controllers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"github.com/SZabrodskii/gophermart-stas/internal/models"
	"github.com/SZabrodskii/gophermart-stas/internal/server"
	"github.com/SZabrodskii/gophermart-stas/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/gopybara/httpbara"
	"github.com/gopybara/httpbara/casual"
	"go.uber.org/fx"
)

type orderControllerDescription struct {
	OrdersAPI   httpbara.Group `group:"/api/user"`
	UploadOrder httpbara.Route `route:"POST /orders" group:"ordersapi" middlewares:"jwt"`
	GetOrders   httpbara.Route `route:"GET /orders" group:"ordersapi" middlewares:"jwt"`
}

type newOrderControllerIn struct {
	fx.In

	Logger       httpbara.Logger
	OrderService domain.OrderServiceI
}

type orderController struct {
	orderControllerDescription

	logger       httpbara.Logger
	orderService domain.OrderServiceI
}

type UploadOrderRequest struct {
	OrderNumber string
}

type OrdersResponse struct {
	Orders []models.Order
}

func (or *OrdersResponse) StatusCode() int {
	if len(or.Orders) == 0 {
		return http.StatusNoContent
	}
	return http.StatusOK
}

type UploadOrderResponse struct {
	Code int
}

func (ur *UploadOrderResponse) StatusCode() int {
	return ur.Code
}

func NewOrderController(in newOrderControllerIn) (server.AsHandlerOut, error) {
	return server.AsHandler(&orderController{
		logger:       in.Logger,
		orderService: in.OrderService,
	})
}

func (oc *orderController) UploadOrder(ctx *gin.Context, _ *UploadOrderRequest) (*UploadOrderResponse, error) {
	userID, ok := GetUserIDFromContext(ctx.Request.Context())
	if !ok {
		oc.logger.Warn("User not authenticated")
		return nil, casual.NewHTTPErrorFromMessage(401, "Unauthorized")
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		oc.logger.Error("Failed to read request body", "error", err)
		return nil, casual.NewHTTPErrorFromMessage(400, "Failed to read request body")
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		oc.logger.Warn("Empty order number")
		return nil, casual.NewHTTPErrorFromMessage(400, "Order number cannot be empty")
	}

	err = oc.orderService.UploadOrder(userID, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidOrderNumber):
			oc.logger.Warn("Invalid order number format", "order_number", orderNumber)
			return nil, casual.NewHTTPErrorFromMessage(422, "Invalid order number format")
		case errors.Is(err, services.ErrOrderExists):
			oc.logger.Info("Order already uploaded by user", "order_number", orderNumber, "user_id", userID)
			return &UploadOrderResponse{Code: http.StatusOK}, nil
		case errors.Is(err, services.ErrOrderExistsOtherUser):
			oc.logger.Warn("Order already uploaded by different user", "order_number", orderNumber)
			return nil, casual.NewHTTPErrorFromMessage(409, "Order already uploaded by different user")
		default:
			oc.logger.Error("Failed to upload order", "error", err, "order_number", orderNumber, "user_id", userID)
			return nil, casual.NewHTTPErrorFromMessage(500, "Failed to upload order")
		}
	}

	oc.logger.Info("Order uploaded successfully", "order_number", orderNumber, "user_id", userID)
	return &UploadOrderResponse{Code: http.StatusAccepted}, nil
}

func (oc *orderController) GetOrders(ctx context.Context, _ *struct{}) (*OrdersResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		oc.logger.Warn("User not authenticated")
		return nil, casual.NewHTTPErrorFromMessage(401, "Unauthorized")
	}

	orders, err := oc.orderService.GetUserOrders(userID)
	if err != nil {
		oc.logger.Error("Failed to get orders", "error", err, "user_id", userID)
		return nil, casual.NewHTTPErrorFromMessage(500, "Failed to get orders")
	}

	oc.logger.Info("Orders retrieved successfully", "count", len(orders), "user_id", userID)
	return &OrdersResponse{Orders: orders}, nil
}
