package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"flypro-assessment-ayo-rt/internal/config"
	"flypro-assessment-ayo-rt/internal/dto"
	"flypro-assessment-ayo-rt/internal/models"
	"flypro-assessment-ayo-rt/internal/repository"
	"flypro-assessment-ayo-rt/internal/utils"

	"github.com/redis/go-redis/v9"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error)
}

type userService struct {
	userRepo    repository.UserRepository
	redisClient *redis.Client
	config      *config.Config
}

func NewUserService(userRepo repository.UserRepository, redisClient *redis.Client, cfg *config.Config) UserService {
	return &userService{
		userRepo:    userRepo,
		redisClient: redisClient,
		config:      cfg,
	}
}

func (s *userService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	existingUser, _ := s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, utils.NewConflictError("user with this email already exists")
	}

	user := &models.User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, utils.NewInternalError("failed to create user", err)
	}

	return s.mapToUserResponse(user), nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error) {
	cacheKey := fmt.Sprintf("user:%d", id)
	cachedUser, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cachedUser), &user); err == nil {
			return s.mapToUserResponse(&user), nil
		}
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, utils.NewNotFoundError("user")
	}

	userJSON, _ := json.Marshal(user)
	s.redisClient.Set(ctx, cacheKey, userJSON, time.Hour)

	return s.mapToUserResponse(user), nil
}

func (s *userService) mapToUserResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}