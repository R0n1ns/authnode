package postgres

import (
	"authmicro/internal/data/models"
	"context"
	"gorm.io/gorm"
	_ "gorm.io/gorm/logger"
	_ "log"
	"time"
)

// Интерфейс репозитория пользователей
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

// Реализация репозитория пользователей с использованием GORM
type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Интерфейс сервиса пользователей
type UserService interface {
	RegisterUser(ctx context.Context, username, email, password, role string) error
	AuthenticateUser(ctx context.Context, email, password string) (*models.User, error)
}

// Реализация сервиса пользователей
type UserServiceImpl struct {
	userRepo UserRepository
	//redis    *redis.Client
}

func NewUserService(userRepo UserRepository) *UserServiceImpl {
	return &UserServiceImpl{userRepo: userRepo}
}

func (s *UserServiceImpl) RegisterUser(ctx context.Context, username, email, password, role string) error {
	// Хеширование пароля и другие проверки можно добавить здесь
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: password, // Замените на хешированный пароль
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	return s.userRepo.CreateUser(ctx, user)
}

func (s *UserServiceImpl) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	// Реализация аутентификации пользователя
	return nil, nil
}

//func main() {
//	// Настройка подключения к PostgreSQL
//	dsn := "host=localhost user=your_user password=your_password dbname=your_db port=5432 sslmode=disable"
//	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
//		Logger: logger.Default.LogMode(logger.Info),
//	})
//	if err != nil {
//		log.Fatalf("Ошибка подключения к базе данных: %v", err)
//	}
//
//	// Автоматическая миграция схемы
//	if err := db.AutoMigrate(&User{}, &Role{}, &VerifyCode{}); err != nil {
//		log.Fatalf("Ошибка миграции схемы: %v", err)
//	}
//
//	// Настройка подключения к Redis
//	redisClient := redis.NewClient(&redis.Options{
//		Addr:     "localhost:6379",
//		Password: "",
//		DB:       0,
//	})
//
//	ctx := context.Background()
//	if _, err := redisClient.Ping(ctx).Result(); err != nil {
//		log.Fatalf("Ошибка подключения к Redis: %v", err)
//	}
//
//	// Создание репозитория и сервиса
//	userRepo := NewGormUserRepository(db)
//	userService := NewUserService(userRepo, redisClient)
//
//	// Пример использования сервиса
//	if err := userService.RegisterUser(ctx, "johndoe", "john@example.com", "securepassword", "user"); err != nil {
//		log.Fatalf("Ошибка регистрации пользователя: %v", err)
//	}
//}
