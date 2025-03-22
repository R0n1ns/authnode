package otp

import (
	"authmicro/internal/data/models"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"
)

// Интерфейс сервиса OTP
type OTPService interface {
	GenerateAndSendOTP(ctx context.Context, email string) error
	VerifyOTP(ctx context.Context, email, code string) (bool, error)
}

// Реализация сервиса OTP
type OTPServiceImpl struct {
	otpRepo      OTPRepository
	emailService EmailService
}

// Конструктор сервиса
func NewOTPService(otpRepo OTPRepository, emailService EmailService) *OTPServiceImpl {
	return &OTPServiceImpl{otpRepo: otpRepo, emailService: emailService}
}

// Генерация случайного OTP-кода (4-6 цифр)
func generateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// Генерация, сохранение и отправка OTP на email
func (s *OTPServiceImpl) GenerateAndSendOTP(ctx context.Context, email string) error {
	otpCode := generateOTP()
	expiresAt := time.Now().Add(5 * time.Minute)

	otp := &models.OTP{
		Email:     email,
		Code:      otpCode,
		ExpiresAt: expiresAt,
	}

	// Сохраняем OTP в БД
	if err := s.otpRepo.SaveOTP(ctx, otp); err != nil {
		return fmt.Errorf("ошибка сохранения OTP: %w", err)
	}

	// Отправляем email
	if err := s.emailService.SendOTPEmail(email, otpCode); err != nil {
		log.Printf("Ошибка отправки email: %v", err)
		return err
	}

	return nil
}

// Проверка OTP-кода
func (s *OTPServiceImpl) VerifyOTP(ctx context.Context, email, code string) (bool, error) {
	otp, err := s.otpRepo.GetOTP(ctx, email)
	if err != nil {
		return false, fmt.Errorf("ошибка получения OTP: %w", err)
	}

	if time.Now().After(otp.ExpiresAt) {
		s.otpRepo.DeleteOTP(ctx, email) // Удаляем просроченный код
		return false, fmt.Errorf("OTP-код просрочен")
	}

	if otp.Code != code {
		return false, fmt.Errorf("неверный OTP-код")
	}

	s.otpRepo.DeleteOTP(ctx, email) // Удаляем использованный код
	return true, nil
}
