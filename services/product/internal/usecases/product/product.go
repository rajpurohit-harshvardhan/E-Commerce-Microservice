package product

import (
	"common/utils/response"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"product/internal/db"
	"product/internal/entities"
	"strconv"

	"github.com/go-playground/validator/v10"
)

func HealthCheck() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write([]byte("Welcome to Product Service!"))
	}
}

func New(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`CreateProduct :: start`)

		var product entities.Product

		err := json.NewDecoder(request.Body).Decode(&product)
		if errors.Is(err, io.EOF) {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body: %w", err)))
			return
		}

		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf(" Error while decoding request body : %w", err)))
			return
		}

		if err := validator.New().Struct(product); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(writer, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		id, err := db.CreateProduct(product.SKU, product.Name, product.Description, product.Price, product.Stock)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf(" Error creating product : %w", err)))
			return
		}

		slog.Info(`CreateProduct :: Product created successfully :`, slog.String("id", id))
		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(map[string]string{"id": id}))
	}
}

func ListProducts(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`ListProducts :: start`)

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
		products, err := db.ListProducts(limit, offset)

		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf(" Error fetching products details : %w", err)))
			return
		}

		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(products))
	}
}

func DeleteProductById(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`DeleteProductById :: start`)

		id := request.PathValue("id")
		if id == "" {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf(" product id is required")))
			return
		}

		ok, err := db.DeleteProductById(id)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf(" Error while decoding request body : %w", err)))
			return
		}

		if !ok {
			response.WriteJson(writer, http.StatusNotFound, response.GeneralError(fmt.Errorf("product not found")))
			return
		}

		response.WriteJson(writer, http.StatusNoContent, response.GeneralResponse(true))
	}
}

func UpdateProductById(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`UpdateProductById :: start`)
		id := request.PathValue("id")
		if id == "" {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf(" product id is required")))
			return
		}

		var product map[string]interface{}
		err := json.NewDecoder(request.Body).Decode(&product)
		if errors.Is(err, io.EOF) {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf(" Error while decoding request body : %w", err)))
			return
		}

		if v, ok := product["price"]; ok {
			if f, ok := v.(float64); ok {
				product["price"] = f // for DECIMAL
			}
		}
		if v, ok := product["stock"]; ok {
			switch t := v.(type) {
			case float64:
				product["stock"] = int64(t) // cast to int64
			case int64:
				product["stock"] = t
			default:
				delete(product, "stock") // if invalid
			}
		}

		result, err := db.UpdateProductById(id, product)
		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf(" Error updating product : %w", err)))
			return
		}

		slog.Info(`UpdateProductById :: Product updated successfully :`, slog.String("id", id))
		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(result))
	}
}

func GetProductById(db db.Db) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info(`GetProductById :: start`)

		id := r.PathValue("id")
		if id == "" {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("product id is required")))
			return
		}
		pr, err := db.GetProductById(id)
		if err != nil {
			response.WriteJson(w, http.StatusNotFound, response.GeneralError(fmt.Errorf("product not found")))
			return
		}
		slog.Info("GetProductById :: Product retrieved successfully :", pr)
		response.WriteJson(w, http.StatusOK, response.GeneralResponse(pr))
	}
}
