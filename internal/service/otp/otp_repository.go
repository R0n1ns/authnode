package otp

import (
	"authmicro/internal/data/models"
	"context"
	_ "time"

	"gorm.io/gorm"
)

// Интерфейс репозитория OTP
type OTPRepository interface {
	SaveOTP(ctx context.Context, otp *models.OTP) error
	GetOTP(ctx context.Context, email string) (*models.OTP, error)
	DeleteOTP(ctx context.Context, email string) error
}

// Реализация репозитория
type GormOTPRepository struct {
	db *gorm.DB
}

// Конструктор репозитория
func NewOTPRepository(db *gorm.DB) *GormOTPRepository {
	return &GormOTPRepository{db: db}
}

// Сохранение OTP в БД
func (r *GormOTPRepository) SaveOTP(ctx context.Context, otp *models.OTP) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

// Получение OTP по email
func (r *GormOTPRepository) GetOTP(ctx context.Context, email string) (*models.OTP, error) {
	var otp models.OTP
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

// Удаление OTP
func (r *GormOTPRepository) DeleteOTP(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).Where("email = ?", email).Delete(&models.OTP{}).Error
}
