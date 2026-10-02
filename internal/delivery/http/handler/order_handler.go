package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	orderUC "symetra-lab-backend-v2/internal/usecase/order"
)

type OrderHandler struct {
	listUC          *orderUC.ListOrdersUseCase
	getUC           *orderUC.GetOrderUseCase
	createUC        *orderUC.CreateOrderUseCase
	updateStatusUC  *orderUC.UpdateOrderStatusUseCase
	updatePaymentUC *orderUC.UpdatePaymentStatusUseCase
	deleteUC        *orderUC.DeleteOrderUseCase
}

func NewOrderHandler(
	listUC *orderUC.ListOrdersUseCase,
	getUC *orderUC.GetOrderUseCase,
	createUC *orderUC.CreateOrderUseCase,
	updateStatusUC *orderUC.UpdateOrderStatusUseCase,
	updatePaymentUC *orderUC.UpdatePaymentStatusUseCase,
	deleteUC *orderUC.DeleteOrderUseCase,
) *OrderHandler {
	return &OrderHandler{
		listUC:          listUC,
		getUC:           getUC,
		createUC:        createUC,
		updateStatusUC:  updateStatusUC,
		updatePaymentUC: updatePaymentUC,
		deleteUC:        deleteUC,
	}
}

func (h *OrderHandler) List(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	orders, err := h.listUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to retrieve orders", err.Error()))
	}

	resp := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		resp[i] = dto.ToOrderResponse(o)
	}

	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *OrderHandler) GetByID(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid order ID", err.Error()))
	}

	o, err := h.getUC.Execute(c.Request().Context(), userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Order not found", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToOrderResponse(o)))
}

func (h *OrderHandler) Create(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.CreateOrderRequest
	if err := c.Bind(&req); err != nil || req.CustomerName == "" || len(req.Items) == 0 {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Customer name and at least 1 item required"))
	}

	itemsInput := make([]orderUC.CreateOrderItemInput, len(req.Items))
	for i, item := range req.Items {
		var pID *uuid.UUID
		if item.ProductID != nil && *item.ProductID != "" {
			if parsed, err := uuid.Parse(*item.ProductID); err == nil {
				pID = &parsed
			}
		}

		var mID *uuid.UUID
		if item.MachineID != nil && *item.MachineID != "" {
			if parsed, err := uuid.Parse(*item.MachineID); err == nil {
				mID = &parsed
			}
		}

		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}

		hpp := 0.0
		if item.HPP != nil {
			hpp = *item.HPP
		}

		weight := 0.0
		if item.WeightGrams != nil {
			weight = *item.WeightGrams
		}

		printTime := 0.0
		if item.PrintTimeHours != nil {
			printTime = *item.PrintTimeHours
		}

		itemsInput[i] = orderUC.CreateOrderItemInput{
			ProductID:      pID,
			ProductName:    item.ProductName,
			Quantity:       qty,
			SellingPrice:   item.SellingPrice,
			HPP:            hpp,
			WeightGrams:    weight,
			PrintTimeHours: printTime,
			MachineID:      mID,
		}
	}

	ord, err := h.createUC.Execute(c.Request().Context(), userID, orderUC.CreateOrderInput{
		CustomerName:    req.CustomerName,
		CustomerContact: req.CustomerContact,
		Notes:           req.Notes,
		Source:          req.Source,
		PaymentStatus:   req.PaymentStatus,
		Items:           itemsInput,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create order", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToOrderResponse(ord)))
}

func (h *OrderHandler) UpdateStatus(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid order ID", err.Error()))
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	status := req.Status
	if status == "" {
		status = req.OrderStatus
	}
	if status == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Status is required"))
	}

	if err := h.updateStatusUC.Execute(c.Request().Context(), userID, id, status); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update status", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"message": "Order status updated successfully",
		"status":  status,
	}))
}

func (h *OrderHandler) UpdatePaymentStatus(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid order ID", err.Error()))
	}

	var req dto.UpdatePaymentStatusRequest
	if err := c.Bind(&req); err != nil || req.PaymentStatus == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Payment status is required"))
	}

	if err := h.updatePaymentUC.Execute(c.Request().Context(), userID, id, req.PaymentStatus); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update payment status", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"message":        "Payment status updated successfully",
		"payment_status": req.PaymentStatus,
	}))
}

func (h *OrderHandler) Delete(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid order ID", err.Error()))
	}

	if err := h.deleteUC.Execute(c.Request().Context(), userID, id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to delete order", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"message": "Order deleted successfully",
	}))
}
