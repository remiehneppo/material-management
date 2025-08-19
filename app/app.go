package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/remiehneppo/material-management/config"
	_ "github.com/remiehneppo/material-management/docs"
	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/internal/handler"
	"github.com/remiehneppo/material-management/internal/logger"
	"github.com/remiehneppo/material-management/internal/middleware"
	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	api         *gin.Engine
	port        string
	database    database.Database
	redisClient *redis.Client
	logger      *logger.Logger
	config      *config.AppConfig
}

func NewApp(cfg *config.AppConfig) *App {

	logger, err := logger.NewLogger(&cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	api := gin.New()
	api.Use(gin.Recovery())
	api.Use(logger.GinLogger())

	// Initialize database
	db := database.NewMongoDatabase(cfg.MongoDB.URI, cfg.MongoDB.Database)
	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("Connecting to database...")
	if err := db.Connect(ctx); err != nil {
		logger.Fatal("error connect to database")
	}
	logger.Info("Database connected successfully")

	redisOpts := &redis.Options{
		Addr: cfg.Redis.URL,
	}
	if cfg.Redis.Username != "" && cfg.Redis.Password != "" {
		redisOpts.Username = cfg.Redis.Username
		redisOpts.Password = cfg.Redis.Password
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("error connect to redis database")
	}
	logger.Info("redis connected successfully")

	return &App{
		api:         api,
		port:        cfg.Port,
		database:    db,
		logger:      logger,
		config:      cfg,
		redisClient: redisClient,
	}
}

func (a *App) Start() error {
	// Initialize Gin

	// Create server
	srv := &http.Server{
		Addr:    ":" + a.port,
		Handler: a.api,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start server
	go func() {
		a.logger.Info("Server starting on port ", a.port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Channel for listening to OS signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking select waiting for server errors or shutdown signals
	select {
	case err := <-serverErrors:
		a.logger.Error("Server error: ", err)
		return err

	case <-shutdown:
		a.logger.Info("Starting graceful shutdown...")

		// Create context with timeout for shutdown operations
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Shutdown the server
		if err := srv.Shutdown(ctx); err != nil {
			a.logger.Error("Server shutdown error: ", err)

			// Force shutdown if graceful shutdown fails
			if err := srv.Close(); err != nil {
				a.logger.Error("Server forced close error: ", err)
				return err
			}
		}

		// Disconnect from database
		a.logger.Info("Disconnecting from database...")
		if err := a.database.Disconnect(ctx); err != nil {
			a.logger.Error("Database disconnect error: ", err)
			return err
		}

		a.logger.Info("Graceful shutdown completed")
	}

	return nil
}

func (a *App) RegisterHandler() {
	userRepo := repository.NewUserRepository(a.database)
	materialsProfileRepo := repository.NewMaterialsProfileRepository(a.database)
	maintenanceRepo := repository.NewMaintenanceRepository(a.database)
	equipmentMachineryRepo := repository.NewEquipmentMachineryRepo(a.database)

	jwtService := service.NewJWTService(
		a.config.JWT.Secret,
		a.config.JWT.Issuer,
		a.config.JWT.Expire,
	)

	uploadService := service.NewUploadService("uploads/")
	
	loginService := service.NewLoginService(jwtService, userRepo)
	materialsProfileService := service.NewMaterialsProfileService(materialsProfileRepo, maintenanceRepo, equipmentMachineryRepo, uploadService)

	loginHandler := handler.NewLoginHandler(loginService, a.logger)
	materialProfileHandler := handler.NewMaterialProfileHandler(materialsProfileService, a.logger)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	a.api.Use(middleware.CorsMiddleware)
	// Register routes

	a.api.Handle("GET", "/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	a.api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	a.api.POST("/api/v1/auth/login", loginHandler.Login)
	a.api.POST("/api/v1/auth/logout", authMiddleware.AuthBearerMiddleware(), loginHandler.Logout)
	a.api.POST("/api/v1/auth/refresh", authMiddleware.AuthBearerMiddleware(), loginHandler.Refresh)

	// Materials Profile routes
	a.api.GET("/api/v1/materials-profile/:id", authMiddleware.AuthBearerMiddleware(), materialProfileHandler.GetMaterialsProfileByID)
	a.api.POST("/api/v1/materials-profile/filter", authMiddleware.AuthBearerMiddleware(), materialProfileHandler.FilterMaterialsProfiles)
	a.api.POST("/api/v1/materials-profile/upload-estimate-sheet", authMiddleware.AuthBearerMiddleware(), materialProfileHandler.UpdateMaterialsEstimateProfileBySheet)

	// Middleware

}
