package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken = fmt.Errorf("invalid token")
	ErrExpiredToken = fmt.Errorf("token has expired")
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type Auth struct {
	*JWTManager
}

type Config struct {
	SecretKey     string        `yaml:"secret_key"`
	TokenDuration time.Duration `yaml:"token_duration"`
}

func New(config Config) (*Auth, error) {
	if config.SecretKey == "" {
		return nil, fmt.Errorf("secret key is required")
	}

	if config.TokenDuration == 0 {
		config.TokenDuration = 24 * time.Hour
	}

	jwtManager := &JWTManager{
		secretKey:     config.SecretKey,
		tokenDuration: config.TokenDuration,
	}

	return &Auth{
		JWTManager: jwtManager,
	}, nil
}

func (m *JWTManager) GenerateToken(userID string, claims map[string]interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(m.tokenDuration).Unix(),
	})

	for key, value := range claims {
		token.Claims.(jwt.MapClaims)[key] = value
	}

	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
} 
