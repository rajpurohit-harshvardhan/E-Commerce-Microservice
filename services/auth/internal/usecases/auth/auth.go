package auth

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
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/golang-jwt/jwt/v5"
)

func Login(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`Login :: start`)

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

		userDetails, err := db.GetUserByEmail(user.Email)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("User does not exist ")))
			return
		}

		same := auth.CheckPasswordHash(user.Password, userDetails.PasswordHash)
		if !same {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("Invalid password: %w", err)))
			return
		}

		accessTokenExpiresAt := time.Now().Add(time.Hour * 2)
		accessToken, err := auth.CreateToken(userDetails.ID, "access", accessTokenExpiresAt)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("Error creating token: %w", err)))
			return
		}

		refreshToken, err := auth.GenerateRefreshToken()
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("Error creating token: %w", err)))
			return
		}
		refreshTokenExpiresAt := time.Now().Add(time.Hour * 24 * 7)
		refreshTokenIssuedAt := time.Now()

		_, err = db.CreateRefreshToken(userDetails.ID, refreshToken.Hash, refreshTokenExpiresAt, refreshTokenIssuedAt, false)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		slog.Info(`Login :: User Logged in successfully :`)
		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(map[string]interface{}{
			"accessToken":  accessToken,
			"refreshToken": refreshToken.Raw,
			"expiresAt":    accessTokenExpiresAt,
		}))
	}
}

func RefreshTokens(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`RefreshTokens :: start`)

		input := make(map[string]string)

		err := json.NewDecoder(request.Body).Decode(&input)
		if errors.Is(err, io.EOF) {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body: %w", err)))
			return
		}

		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if input["refreshToken"] == "" {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf(" refresh token is required: %w", err)))
			return
		}

		hash := auth.HashRefreshToken(input["refreshToken"])
		token, err := db.GetTokenByHash(hash)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("Error when fetching token details: %w ", err)))
			return
		}

		// checking whether the token is expired or not
		if time.Now().After(token.ExpiresAt) {
			response.WriteJson(writer, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("Token expired: %w", err)))
			return
		}

		// deleting the old token from table
		ok, err := db.DeleteTokenById(token.ID)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("Error when deleting token details: %w ", err)))
			return
		}

		if !ok {
			response.WriteJson(writer, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("No token deleted: %w ", err)))
			return
		}

		// creating new tokens and storing them in table
		accessTokenExpiresAt := time.Now().Add(time.Hour * 2)
		accessToken, err := auth.CreateToken(token.UserId, "access", accessTokenExpiresAt)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("Error creating token: %w", err)))
			return
		}

		refreshToken, err := auth.GenerateRefreshToken()
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("Error creating token: %w ", err)))
			return
		}
		refreshTokenExpiresAt := time.Now().Add(time.Hour * 24 * 7)
		refreshTokenIssuedAt := time.Now()

		_, err = db.CreateRefreshToken(token.UserId, refreshToken.Hash, refreshTokenExpiresAt, refreshTokenIssuedAt, false)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("Error creating token: %w ", err)))
			return
		}

		slog.Info(`RefreshTokens :: End :`)
		response.WriteJson(writer, http.StatusOK, response.GeneralResponse(map[string]interface{}{
			"accessToken":  accessToken,
			"refreshToken": refreshToken.Raw,
			"expiresAt":    accessTokenExpiresAt,
		}))
	}
}

func Logout(db db.Db) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		slog.Info(`Logout :: start`)

		input := make(map[string]string)

		err := json.NewDecoder(request.Body).Decode(&input)
		if errors.Is(err, io.EOF) {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body: %w", err)))
			return
		}

		if err != nil {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if input["refreshToken"] == "" {
			response.WriteJson(writer, http.StatusBadRequest, response.GeneralError(fmt.Errorf(" refresh token is required: %w", err)))
			return
		}

		hash := auth.HashRefreshToken(input["refreshToken"])
		// deleting the token by userId from table
		ok, err := db.DeleteTokenByHash(hash)
		if err != nil {
			response.WriteJson(writer, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		if !ok {
			response.WriteJson(writer, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("No token deleted: %w ", err)))
			return
		}

		slog.Info(`Logout :: End :`)
		response.WriteJson(writer, http.StatusNoContent, "")
	}
}
