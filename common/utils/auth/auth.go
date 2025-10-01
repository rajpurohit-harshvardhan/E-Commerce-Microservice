package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const secretKeyString = "5bk47N6DVwEELnzv0kYr/xIIa3VSdSPDAJpTf2zPyDA="

var (
	secretKey []byte
	once      sync.Once
)

type UserClaims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type RefreshToken struct {
	Raw  string // send to client
	Hash string // store in DB: hex(sha256(Raw))
}

func CreateHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetSecretKey() []byte {
	once.Do(func() {
		key := os.Getenv("JWT_SECRET_KEY")
		if key == "" {
			key = secretKeyString
			//panic("FATAL: JWT Secret Key not defined.")
		}
		secretKey = []byte(key)
	})
	return secretKey
}

func CreateToken(id string, tokenType string, exp time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":  id,
			"type": tokenType,
			"exp":  exp.Unix(),
			"iat":  time.Now().Unix(),
			//"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(GetSecretKey())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return GetSecretKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid or expired token")
}

func GenerateRefreshToken() (RefreshToken, error) {
	buf := make([]byte, 32) // 32 bytes -> 64 hex chars
	if _, err := rand.Read(buf); err != nil {
		return RefreshToken{}, err
	}
	raw := hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(raw))
	return RefreshToken{
		Raw:  raw,
		Hash: hex.EncodeToString(sum[:]),
	}, nil
}

func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func ConstantTimeEqualHex(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

var ErrMissingToken = errors.New("missing refresh token")
