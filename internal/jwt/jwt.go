package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Конфигурация для JWT
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// Структура сервиса JWT
type JWTService struct {
	config JWTConfig
}

// Структура для хранения claims в токене
type TokenClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Создание нового JWT-сервиса
func NewJWTService(config JWTConfig) *JWTService {
	return &JWTService{config: config}
}

// Генерация пары access и refresh токенов
func (s *JWTService) GenerateTokenPair(userID uint, role string) (accessToken string, refreshToken string, err error) {
	// Создание access-токена
	accessToken, err = s.generateToken(userID, role, s.config.AccessTTL, s.config.AccessSecret)
	if err != nil {
		return "", "", fmt.Errorf("ошибка генерации access-токена: %w", err)
	}

	// Создание refresh-токена
	refreshToken, err = s.generateToken(userID, role, s.config.RefreshTTL, s.config.RefreshSecret)
	if err != nil {
		return "", "", fmt.Errorf("ошибка генерации refresh-токена: %w", err)
	}

	return accessToken, refreshToken, nil
}

// Проверка и парсинг токена
func (s *JWTService) ValidateToken(tokenString string, isRefresh bool) (*TokenClaims, error) {
	secret := s.config.AccessSecret
	if isRefresh {
		secret = s.config.RefreshSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неподдерживаемый метод подписи")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка валидации токена: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("невалидный токен")
	}

	return claims, nil
}

// Обновление пары токенов
func (s *JWTService) RefreshToken(refreshToken string) (string, string, error) {
	// Проверяем refresh-токен
	claims, err := s.ValidateToken(refreshToken, true)
	if err != nil {
		return "", "", fmt.Errorf("ошибка валидации refresh-токена: %w", err)
	}

	// Генерируем новую пару
	return s.GenerateTokenPair(claims.UserID, claims.Role)
}

// Внутренний метод для генерации JWT-токена
func (s *JWTService) generateToken(userID uint, role string, ttl time.Duration, secret string) (string, error) {
	expirationTime := time.Now().Add(ttl)

	claims := &TokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
