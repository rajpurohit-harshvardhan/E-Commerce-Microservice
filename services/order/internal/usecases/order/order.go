package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"order/internal/entities"
	"strconv"

	"log/slog"

	"net/http"
	"order/internal/db"

	"common/utils/response"
	//"order/internal/utils/response"

	"github.com/go-playground/validator/v10"
)

func HealthCheck() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write([]byte("Welcome to Order Service!"))
	}
}

func New(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`CreateOrder :: start`)

		userId := request.Header.Get("x-user-id")
		if userId == "" {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("Missing User Id")))
			return
		}

		var order entities.CreateOrderRequestInput

		err := json.NewDecoder(request.Body).Decode(&order)
		if errors.Is(err, io.EOF) {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if err := validator.New().Struct(order); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(writer, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		total := 0.0
		for _, orderItem := range order.Items {
			total += orderItem.Price * float64(orderItem.Quantity)
		}

		id, err := db.CreateOrder(userId, "CONFIRMED", total)
		if err != nil {
			slog.Error("Error while creating auth: ", err)
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		var orderItems []entities.OrderItem
		for _, orderItem := range order.Items {
			orderItems = append(orderItems, entities.OrderItem{
				OrderId:   id,
				ProductId: orderItem.ID,
				Quantity:  orderItem.Quantity,
				Price:     orderItem.Price,
			})
		}

		result, err := db.CreateOrderItems(orderItems)
		if err != nil {
			slog.Error("Error while creating auth items: ", err)
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		if !result {
			slog.Error("Error : No auth items provided: ", err)
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf(`no auth items provided`)))
			return
		}

		slog.Info(`CreateOrder :: Order created successfully :`, slog.String("id", id))
		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(map[string]string{"id": id}))
	}
}

func ListOrders(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`ListOrders :: start`)

		queryParams := request.URL.Query()
		limit := 10
		if v := queryParams.Get("limit"); v != "" {
			if n, _ := strconv.Atoi(v); n > 0 && n <= 100 {
				limit = n
			}
		}
		offset := 0
		if v := queryParams.Get("offset"); v != "" {
			if n, _ := strconv.Atoi(v); n >= 0 {
				offset = n
			}
		}

		orders, err := db.ListOrders(limit, offset)

		if err != nil {
			slog.Error("error getting orders", err.Error())
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		slog.Info(`ListOrders :: End`)
		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(orders))
	}
}

func DeleteOrderById(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`DeleteOrderById :: start`)

		id := request.PathValue("id")
		if id == "" {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(nil))
			return
		}

		ok, err := db.DeleteOrderById(id)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		if !ok {
			response.WriteJson(writer, http.StatusNotFound, response.GeneralError(fmt.Errorf("auth not found")))
			return
		}

		slog.Info(`DeleteOrderById :: Order Deleted Successfully`)
		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(ok))
	}
}

func GetOrderById(db db.Db) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info(`GetOrderById :: Start`)
		id := r.PathValue("id")
		if id == "" {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("missing id")))
			return
		}

		var orderDetails entities.OrderDetails

		order, err := db.GetOrderById(id)
		if err != nil {
			response.WriteJson(w, http.StatusNotFound, response.GeneralError(fmt.Errorf("auth not found")))
			return
		}

		orderItems, err := db.GetOrderItemsByOrderId(id)
		if err != nil {
			response.WriteJson(w, http.StatusNotFound, response.GeneralError(fmt.Errorf("Order items not found")))
			return
		}

		orderDetails.Order = order
		orderDetails.Items = orderItems

		slog.Info(`GetOrderById :: End`)
		response.WriteJson(w, http.StatusOK, response.GeneralResponse(orderDetails))
	}
}

func UpdateOrderById(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`UpdateOrderById :: start`)
		id := request.PathValue("id")
		if id == "" {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(nil))
		}

		var order map[string]interface{}
		err := json.NewDecoder(request.Body).Decode(&order)
		if errors.Is(err, io.EOF) {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		var detailsToUpdate map[string]interface{}
		if v, ok := order["status"]; ok {
			detailsToUpdate["status"] = v
		}

		type OrderItemsToUpdate = entities.CreateOrderRequestInputItems
		var OrderItemsToCreate []entities.OrderItem
		if v, ok := order["items"]; ok {
			// calculate total based on the current items in the body
			if items, ok := v.([]OrderItemsToUpdate); ok {
				total := 0.0
				for _, orderItem := range items {
					total += orderItem.Price * float64(orderItem.Quantity)
					OrderItemsToCreate = append(OrderItemsToCreate, entities.OrderItem{
						OrderId:   id,
						ProductId: orderItem.ID,
						Quantity:  orderItem.Quantity,
						Price:     orderItem.Price,
					})
				}

				detailsToUpdate["total"] = total

				// delete all the auth items
				_, err := db.DeleteOrderItemsByOrderId(id)
				if err != nil {
					response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
					return
				}

				// create all the auth items
				_, err = db.CreateOrderItems(OrderItemsToCreate)
				if err != nil {
					response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
					return
				}

			} else {
				// log or return error
				slog.Info("No auth items provided. Only updating status")
			}
		}

		result, err := db.UpdateOrderById(id, detailsToUpdate)
		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, map[string]interface{}{"error": err.Error(), "result": result})
			return
		}

		slog.Info(`UpdateOrderById :: Order updated successfully :`, slog.String("id", id))
		response.WriteJson(writer, http.StatusOK, map[string]interface{}{"id": id, "result": result})
	}
}
