package user

import (
	"auth/internal/db"
	"auth/internal/entities"
	"common/utils/auth"
	"common/utils/response"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
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
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body: %w", err)))
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

		passwordHash, err := auth.CreateHash(user.Password)
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

		slog.Info(`CreateUser :: User created successfully :`, slog.String("id", id))
		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(map[string]string{"id": id}))
	}
}

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
			response.WriteJson(writer, http.StatusNotFound, response.GeneralError(fmt.Errorf("user not found: %w", err)))
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
			response.WriteJson(w, http.StatusNotFound, response.GeneralError(fmt.Errorf("user not found: %w", err)))
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
			return
		}

		var user map[string]interface{}
		err := json.NewDecoder(request.Body).Decode(&user)
		if errors.Is(err, io.EOF) {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body: %w", err)))
			return
		}

		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		detailsToUpdate := make(map[string]interface{})
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

			hashPassword, err1 := auth.CreateHash(passwordString)
			if err1 != nil {
				response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err1))
				return
			}
			detailsToUpdate["password_hash"] = hashPassword
		}

		result, err := db.UpdateUserById(id, detailsToUpdate)
		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, map[string]interface{}{"error": err.Error(), "result": result})
			return
		}

		slog.Info(`UpdateUserById :: user updated successfully :`, slog.String("id", id))
		response.WriteJson(writer, http.StatusOK, map[string]interface{}{"id": id, "result": result})
	}
}
