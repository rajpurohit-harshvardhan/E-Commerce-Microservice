package user

import (
	"auth/internal/db"
	"auth/internal/entities"
	"common/utils/response"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

func HealthCheck() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write([]byte("Welcome to Auth Service!"))
	}
}

func New(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`CreateUser :: start`)

		var user entities.CreateUserRequest

		err := json.NewDecoder(request.Body).Decode(&user)
		if errors.Is(err, io.EOF) {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if err := validator.New().Struct(user); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(writer, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		// TODO: code to convert password into hash
		passwordHash, err := HashPassword(user.Password)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		id, err := db.CreateUser(user.Name, user.Email, passwordHash)
		if err != nil {
			slog.Error("Error while creating user: ", err)
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		slog.Info(`CreateOrder :: User created successfully :`, slog.String("id", id))
		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(map[string]string{"id": id}))
	}
}

//func ListOrders(db db.Db) http.HandlerFunc {
//	return func(writer http.ResponseWriter, request *http.Request) {
//		slog.Info(`ListOrders :: start`)
//
//		queryParams := request.URL.Query()
//		limit := 10
//		if v := queryParams.Get("limit"); v != "" {
//			if n, _ := strconv.Atoi(v); n > 0 && n <= 100 {
//				limit = n
//			}
//		}
//		offset := 0
//		if v := queryParams.Get("offset"); v != "" {
//			if n, _ := strconv.Atoi(v); n >= 0 {
//				offset = n
//			}
//		}
//
//		orders, err := db.ListOrders(limit, offset)
//
//		if err != nil {
//			slog.Error("error getting orders", err.Error())
//			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
//			return
//		}
//		slog.Info(`ListOrders :: End`)
//		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(orders))
//	}
//}

func DeleteUserById(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`DeleteUserById :: start`)

		id := request.PathValue("id")
		if id == "" {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(nil))
			return
		}

		ok, err := db.DeleteUserById(id)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		if !ok {
			response.WriteJson(writer, http.StatusNotFound, response.GeneralError(fmt.Errorf("user not found")))
			return
		}

		slog.Info(`DeleteUserById :: User Deleted Successfully`)
		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(ok))
	}
}

func GetUserById(db db.Db) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info(`GetUserById :: Start`)
		id := r.PathValue("id")
		if id == "" {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("missing id")))
			return
		}

		var user entities.User

		user, err := db.GetUserById(id)
		if err != nil {
			response.WriteJson(w, http.StatusNotFound, response.GeneralError(fmt.Errorf("user not found")))
			return
		}

		user.PasswordHash = ""

		slog.Info(`GetUserById :: End`)
		response.WriteJson(w, http.StatusOK, response.GeneralResponse(user))
	}
}

func UpdateUserById(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`UpdateUserById :: start`)
		id := request.PathValue("id")
		if id == "" {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(nil))
		}

		var user map[string]interface{}
		err := json.NewDecoder(request.Body).Decode(&user)
		if errors.Is(err, io.EOF) {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		var detailsToUpdate map[string]interface{}
		if v, ok := user["email"]; ok {
			detailsToUpdate["email"] = v
		}

		if v, ok := user["name"]; ok {
			detailsToUpdate["name"] = v
		}

		if v, ok := user["password"]; ok {
			passwordString, ok := v.(string)
			if !ok {
				response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("password must be a string")))
				return
			}

			if passwordString == "" {
				response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("password cannot be empty")))
				return
			}

			hashPassword, err1 := HashPassword(passwordString)
			if err1 != nil {
				response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err1))
				return
			}
			detailsToUpdate["passwordHash"] = hashPassword
		}

		result, err := db.UpdateUserById(id, detailsToUpdate)
		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, map[string]interface{}{"error": err.Error(), "result": result})
			return
		}

		slog.Info(`UpdateUserById :: Order updated successfully :`, slog.String("id", id))
		response.WriteJson(writer, http.StatusOK, map[string]interface{}{"id": id, "result": result})
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
