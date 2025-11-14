package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flypro-assessment-ayo-rt/internal/config"
	"flypro-assessment-ayo-rt/internal/handler"
	"flypro-assessment-ayo-rt/internal/middleware"
	"flypro-assessment-ayo-rt/internal/repository"
	"flypro-assessment-ayo-rt/internal/services"
	"flypro-assessment-ayo-rt/internal/validators"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := initLogger(cfg.App.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	db, err := initDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	redisClient := initRedis(cfg)
	defer redisClient.Close()

	validator := validator.New()
	if err := validators.RegisterCustomValidators(validator); err != nil {
		logger.Fatal("Failed to register validators", zap.Error(err))
	}

	userRepo := repository.NewUserRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	reportRepo := repository.NewExpenseReportRepository(db)

	currencyService := services.NewCurrencyService(redisClient, cfg)
	userService := services.NewUserService(userRepo, redisClient, cfg)
	expenseService := services.NewExpenseService(expenseRepo, currencyService, redisClient, cfg)
	reportService := services.NewExpenseReportService(reportRepo, expenseRepo, currencyService, redisClient, cfg)

	userHandler := handler.NewUserHandler(userService, validator)
	expenseHandler := handler.NewExpenseHandler(expenseService, validator)
	reportHandler := handler.NewExpenseReportHandler(reportService, validator)

	router := setupRouter(logger, userHandler, expenseHandler, reportHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	go func() {
		logger.Info("Starting server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func initLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}

func setupRouter(
	logger *zap.Logger,
	userHandler *handler.UserHandler,
	expenseHandler *handler.ExpenseHandler,
	reportHandler *handler.ExpenseReportHandler,
) *gin.Engine {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler(logger))

	api := router.Group("/api")
	{
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
		}

		expenses := api.Group("/expenses")
		{
			expenses.POST("", expenseHandler.CreateExpense)
			expenses.GET("", expenseHandler.ListExpenses)
			expenses.GET("/:id", expenseHandler.GetExpense)
			expenses.PUT("/:id", expenseHandler.UpdateExpense)
			expenses.DELETE("/:id", expenseHandler.DeleteExpense)
		}

		reports := api.Group("/reports")
		{
			reports.POST("", reportHandler.CreateReport)
			reports.GET("", reportHandler.ListReports)
			reports.GET("/:id", reportHandler.GetReport)
			reports.POST("/:id/expenses", reportHandler.AddExpensesToReport)
			reports.PUT("/:id/submit", reportHandler.SubmitReport)
		}
	}

	return router
}
