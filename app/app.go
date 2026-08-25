package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/remiehneppo/material-management/config"
	_ "github.com/remiehneppo/material-management/docs"
	"github.com/remiehneppo/material-management/internal/database"
	domainequipment "github.com/remiehneppo/material-management/internal/domain/equipment"
	domainmaintenance "github.com/remiehneppo/material-management/internal/domain/maintenance"
	"github.com/remiehneppo/material-management/internal/domain/materialprofile"
	"github.com/remiehneppo/material-management/internal/domain/materialrequest"
	domainsession "github.com/remiehneppo/material-management/internal/domain/session"
	domainuser "github.com/remiehneppo/material-management/internal/domain/user"
	"github.com/remiehneppo/material-management/internal/handler"
	"github.com/remiehneppo/material-management/internal/logger"
	"github.com/remiehneppo/material-management/internal/middleware"
	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/internal/service"
	"github.com/remiehneppo/material-management/types"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	api         *gin.Engine
	port        string
	database    *database.MongoDatabase
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
	created, err := domainuser.BootstrapAdmin(ctx, db.DB(), domainuser.BootstrapAdminConfig{
		Username:  cfg.BootstrapAdmin.Username,
		Password:  cfg.BootstrapAdmin.Password,
		FullName:  cfg.BootstrapAdmin.FullName,
		Workspace: cfg.BootstrapAdmin.Workspace,
	})
	if err != nil {
		logger.Fatal("bootstrap admin: ", err)
	}
	if created {
		logger.Info("Bootstrap admin created successfully")
	}

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
	materialsProfileRepo := repository.NewMaterialsProfileRepository(a.database)
	maintenanceRepo := repository.NewMaintenanceRepository(a.database)
	equipmentMachineryRepo := repository.NewEquipmentMachineryRepo(a.database)
	materialsRequestRepo := repository.NewMaterialsRequestRepository(a.database)

	jwtService := service.NewJWTService(
		a.config.JWT.Secret,
		a.config.JWT.Issuer,
		a.config.JWT.Expire,
	)

	uploadService := service.NewUploadService(a.config.Upload.BaseDir)

	materialsProfileService := service.NewMaterialsProfileService(materialsProfileRepo, maintenanceRepo, equipmentMachineryRepo, uploadService)
	materialsRequestService := service.NewMaterialsRequestService(
		materialsRequestRepo,
		materialsProfileRepo,
		maintenanceRepo,
		equipmentMachineryRepo,
		a.config.MaterialsRequestConfig.TemplatePath,
	)
	materialRequestIssuer := materialrequest.NewIssuer(a.database.Client(), a.database.DB())
	if err := materialRequestIssuer.EnsureIndexes(context.Background()); err != nil {
		a.logger.Fatal("create material request indexes: ", err)
	}
	sessionManager := domainsession.NewManager(a.database.DB(), jwtService, 7*24*time.Hour)
	sessionHandler := domainsession.NewHandler(sessionManager, a.config.Environment == "production")
	userHandler := domainuser.NewHandler(a.database.Client(), a.database.DB())
	materialProfileCatalog := materialprofile.NewCatalog(a.database.DB())
	materialProfileImporter := materialprofile.NewImporter(a.database.Client(), a.database.DB())
	if err := materialProfileCatalog.EnsureIndexes(context.Background()); err != nil {
		a.logger.Fatal("create material profile indexes: ", err)
	}
	materialProfileHandler := handler.NewMaterialProfileHandler(materialsProfileService, materialProfileCatalog, materialProfileImporter, a.logger)
	materialsRequestHandler := handler.NewMaterialRequestHandler(materialsRequestService, materialRequestIssuer, a.logger)
	maintenanceHandler := domainmaintenance.NewHandler(a.database.DB())
	equipmentMachineryHandler := domainequipment.NewHandler(a.database.DB())

	authMiddleware := middleware.NewAuthMiddleware(jwtService, sessionManager)

	allowedOrigins := a.config.CORS.AllowedOrigins
	if len(allowedOrigins) == 0 && a.config.Environment != "production" {
		allowedOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}
	a.api.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// Register routes

	a.api.Handle("GET", "/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	a.api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, types.Response{
			Status: true,
		})
	})

	a.api.POST("/api/v1/auth/login", sessionHandler.Login)
	a.api.POST("/api/v1/auth/refresh", sessionHandler.Refresh)
	a.api.POST("/api/v1/auth/logout", sessionHandler.Logout)

	// User
	userGroup := a.api.Group("/api/v1/user")
	userGroup.Use(authMiddleware.AuthBearerMiddleware())
	userGroup.GET("/profile", userHandler.GetProfile)
	userGroup.POST("/profile", userHandler.UpdateProfile)
	userGroup.POST("/change-password", userHandler.ChangePassword)

	// Maintenance
	maintenanceGroup := a.api.Group("/api/v1/maintenance")
	maintenanceGroup.Use(authMiddleware.AuthBearerMiddleware())
	maintenanceGroup.GET("/:id", maintenanceHandler.Get)
	maintenanceGroup.POST("/filter", maintenanceHandler.Filter)
	maintenanceGroup.POST("/", maintenanceHandler.Create)

	// EquipmentMachinery
	equipmentMachineryGroup := a.api.Group("/api/v1/equipment-machinery")
	equipmentMachineryGroup.Use(authMiddleware.AuthBearerMiddleware())
	equipmentMachineryGroup.POST("/filter", equipmentMachineryHandler.Filter)
	equipmentMachineryGroup.POST("", equipmentMachineryHandler.Create)

	// Materials Profile routes
	materialsProfileGroup := a.api.Group("/api/v1/materials-profiles")
	materialsProfileGroup.Use(authMiddleware.AuthBearerMiddleware())
	materialsProfileGroup.GET("/:id", materialProfileHandler.GetMaterialsProfileByID)
	materialsProfileGroup.POST("/", materialProfileHandler.FilterMaterialsProfiles)
	materialsProfileGroup.GET("/paginated", materialProfileHandler.PaginatedMaterialsProfiles)
	materialsProfileGroup.POST("/upload-estimate", materialProfileHandler.UpdateMaterialsEstimateProfileBySheet)
	materialsProfileGroup.POST("/create", materialProfileHandler.CreateNewMaterialsProfile)
	materialsProfileGroup.POST("/:id/materials", materialProfileHandler.UpsertEstimatedMaterial)

	// Materials Request routes
	materialsRequestGroup := a.api.Group("/api/v1/materials-request")
	materialsRequestGroup.Use(authMiddleware.AuthBearerMiddleware())
	materialsRequestGroup.GET("/:id", materialsRequestHandler.GetMaterialRequestByID)
	materialsRequestGroup.POST("/filter", materialsRequestHandler.FilterMaterialRequests)
	materialsRequestGroup.POST("/export", materialsRequestHandler.ExportMaterialsRequest)
	materialsRequestGroup.POST("/:id/issue", materialsRequestHandler.IssueMaterialRequest)
	materialsRequestGroup.POST("/", materialsRequestHandler.CreateMaterialRequest)
	materialsRequestGroup.POST("/update", materialsRequestHandler.UpdateMaterialRequest)
	materialsRequestGroup.POST("/cancel/:id", materialsRequestHandler.CancelMaterialRequest)

}
