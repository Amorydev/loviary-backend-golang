package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "loviary.app/backend/docs" // Swagger docs - required for UI to load

	"loviary.app/backend/internal/application/auth"
	appChapters "loviary.app/backend/internal/application/chapters"
	"loviary.app/backend/internal/application/couples"
	"loviary.app/backend/internal/application/home"
	"loviary.app/backend/internal/application/memories"
	"loviary.app/backend/internal/application/moods"
	"loviary.app/backend/internal/application/oauth"
	"loviary.app/backend/internal/application/reminders"
	appSparks "loviary.app/backend/internal/application/sparks"
	"loviary.app/backend/internal/application/streaks"
	"loviary.app/backend/internal/application/users"
	"loviary.app/backend/internal/application/verification"
	"loviary.app/backend/internal/infrastructure/email"
	"loviary.app/backend/internal/infrastructure/jwt"
	"loviary.app/backend/internal/infrastructure/persistence/postgres"
	"loviary.app/backend/internal/interfaces/http/handlers"
	"loviary.app/backend/internal/interfaces/http/middleware"
	"loviary.app/backend/pkg/config"
	"loviary.app/backend/pkg/db"
	"loviary.app/backend/pkg/logger"
)

// @title           Loviary API
// @version         1.0
// @description     Loviary - Couple Relationship Tracking API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@loviary.app

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.MustLoad()

	// Initialize logger
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	defer log.Close()

	// Initialize database
	dbConn, err := db.New(&db.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Name:     cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
		Timezone: cfg.Database.Timezone,
	})
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer dbConn.Close()

	// Run migrations
	log.Info("Running database migrations")
	if err := migrate(dbConn.DB.DB, log); err != nil {
		log.Fatal("Failed to run migrations", err)
	}

	// Initialize repositories
	userRepo := persistence.NewUserRepository(dbConn.DB)
	coupleRepo := persistence.NewCoupleRepository(dbConn.DB)
	tokenRepo := persistence.NewRefreshTokenRepository(dbConn.DB)
	moodRepo := persistence.NewMoodRepository(dbConn.DB.DB)
	streakRepo := persistence.NewStreakRepository(dbConn.DB.DB)
	reminderRepo := persistence.NewReminderRepository(dbConn.DB.DB)
	memoryRepo := persistence.NewMemoryRepository(dbConn.DB.DB)
	verificationRepo := persistence.NewEmailVerificationRepository(dbConn.DB)
	chapterRepo := persistence.NewChapterRepository(dbConn.DB.DB)
	sparkRepo := persistence.NewSparkRepository(dbConn.DB.DB)

	// Initialize JWT manager
	jwtManager := jwt.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		cfg.JWT.Issuer,
		cfg.JWT.Audience,
	)

	// Initialize email sender
	emailCfg := &email.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		Enabled:  cfg.SMTP.Enabled,
	}
	emailSender := email.NewSender(emailCfg, log)

	// Log SMTP configuration (hide password)
	log.Info("SMTP Configuration", map[string]interface{}{
		"host":    cfg.SMTP.Host,
		"port":    cfg.SMTP.Port,
		"username": cfg.SMTP.Username,
		"from":    cfg.SMTP.From,
		"enabled":  cfg.SMTP.Enabled,
	})

	// Initialize verification service
	verificationService := verification.NewService(
		verificationRepo,
		emailSender,
		log,
		15*time.Minute,   // code TTL
		1*time.Minute,    // resend window
	)

	// Initialize services
	userService := users.NewService(userRepo)
	coupleService := couples.NewService(coupleRepo)
	authService := auth.NewService(
		userRepo,
		tokenRepo,
		jwtManager,
		verificationService,
		log,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)
	moodService := moods.NewService(moodRepo)
	streakService := streaks.NewService(streakRepo)
	reminderService := reminders.NewService(reminderRepo)
	memoryService := memories.NewService(memoryRepo)
	chapterService := appChapters.NewService(chapterRepo)
	sparkService := appSparks.NewService(sparkRepo)
	homeService := home.NewService(
		userService,
		coupleService,
		moodService,
		streakService,
		chapterService,
		sparkService,
	)

	// Initialize OAuth service
	oauthService := oauth.NewService(
		userRepo,
		tokenRepo,
		jwtManager,
		cfg.OAuth.GoogleClientID,
		cfg.OAuth.GoogleClientSecret,
		cfg.OAuth.GoogleRedirectURI,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(authService, verificationService)
	oauthHandler := handlers.NewOAuthHandler(oauthService)
	coupleHandler := handlers.NewCoupleHandler(coupleService, userService)
	moodHandler := handlers.NewMoodHandler(moodService, streakService)
	streakHandler := handlers.NewStreakHandler(streakService)
	reminderHandler := handlers.NewReminderHandler(reminderService)
	memoryHandler := handlers.NewMemoryHandler(memoryService)
	homeHandler := handlers.NewHomeHandler(homeService)
	storageHandler := handlers.NewStorageHandler()

	// Setup router
	router := gin.New()
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.ErrorMiddleware())
	router.Use(middleware.CORSMiddleware(cfg.CORS.AllowOrigins))

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": cfg.App.Name,
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/resend-verification", authHandler.ResendVerification)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/logout-all", middleware.AuthMiddleware(jwtManager), authHandler.LogoutAll)

			// OAuth routes
			auth.GET("/google", oauthHandler.GoogleRedirect)
			auth.GET("/google/callback", oauthHandler.GoogleCallback)
			auth.POST("/google/mobile", oauthHandler.GoogleMobile)
		}

		// User routes
		users := v1.Group("/users")
		{
			users.POST("/register", userHandler.CreateUser)
			users.GET("/me", middleware.AuthMiddleware(jwtManager), userHandler.GetMyProfile)
			users.GET("/:id", userHandler.GetUser)
			users.PATCH("/:id", middleware.AuthMiddleware(jwtManager), userHandler.UpdateUser)
			users.PATCH("/me", middleware.AuthMiddleware(jwtManager), userHandler.UpdateMyProfile)
			users.PUT("/me/device", middleware.AuthMiddleware(jwtManager), userHandler.UpdateMyDevice)
		}

		// Couple routes
		couples := v1.Group("/couples")
		{
			couples.POST("/invite", middleware.AuthMiddleware(jwtManager), coupleHandler.Invite)
			couples.GET("/me", middleware.AuthMiddleware(jwtManager), coupleHandler.GetMyCouple)
			couples.POST("/confirm", middleware.AuthMiddleware(jwtManager), coupleHandler.ConfirmInvitation)
			couples.DELETE("/me", middleware.AuthMiddleware(jwtManager), coupleHandler.DeleteMyCouple)
		}

		// Mood routes
		moods := v1.Group("/moods")
		{
			moods.POST("", middleware.AuthMiddleware(jwtManager), moodHandler.CreateMood)
			moods.GET("/today", middleware.AuthMiddleware(jwtManager), moodHandler.GetTodaysMood)
			moods.GET("/history", middleware.AuthMiddleware(jwtManager), moodHandler.GetMoodHistory)
			moods.GET("/:id", middleware.AuthMiddleware(jwtManager), moodHandler.GetMood)
			moods.PATCH("/:id", middleware.AuthMiddleware(jwtManager), moodHandler.UpdateMood)
			moods.DELETE("/:id", middleware.AuthMiddleware(jwtManager), moodHandler.DeleteMood)
		}

		// Streak routes
		streaks := v1.Group("/streaks")
		{
			streaks.GET("/me", middleware.AuthMiddleware(jwtManager), streakHandler.GetMyStreaks)
			streaks.GET("/:activity_type", middleware.AuthMiddleware(jwtManager), streakHandler.GetStreak)
			streaks.POST("/:activity_type/record", middleware.AuthMiddleware(jwtManager), streakHandler.RecordActivity)
		}

		// Reminder routes
		reminders := v1.Group("/reminders")
		{
			reminders.GET("", middleware.AuthMiddleware(jwtManager), reminderHandler.GetReminders)
			reminders.POST("", middleware.AuthMiddleware(jwtManager), reminderHandler.CreateReminder)
			reminders.GET("/:id", middleware.AuthMiddleware(jwtManager), reminderHandler.GetReminder)
			reminders.PATCH("/:id", middleware.AuthMiddleware(jwtManager), reminderHandler.UpdateReminder)
			reminders.DELETE("/:id", middleware.AuthMiddleware(jwtManager), reminderHandler.DeleteReminder)
		}

		// Memory routes
		memories := v1.Group("/memories")
		{
			memories.GET("", middleware.AuthMiddleware(jwtManager), memoryHandler.GetMemories)
			memories.POST("", middleware.AuthMiddleware(jwtManager), memoryHandler.CreateMemory)
			memories.GET("/:id", middleware.AuthMiddleware(jwtManager), memoryHandler.GetMemory)
			memories.PATCH("/:id", middleware.AuthMiddleware(jwtManager), memoryHandler.UpdateMemory)
			memories.DELETE("/:id", middleware.AuthMiddleware(jwtManager), memoryHandler.DeleteMemory)
		}

		// Home routes
		home := v1.Group("/home")
		{
			home.GET("", middleware.AuthMiddleware(jwtManager), homeHandler.GetDashboard)
		}

		// Storage routes
		storage := v1.Group("/storage")
		{
			storage.POST("/presign", middleware.AuthMiddleware(jwtManager), storageHandler.Presign)
		}
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port)
	log.Info("About to start server", map[string]interface{}{"addr": addr})

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	log.Info("Server configured, starting ListenAndServe...")
	// Start server in goroutine
	go func() {
		log.Info("ListenAndServe starting...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error", err)
		}
		log.Info("ListenAndServe exited normally")
	}()

	log.Info("Server goroutine started, waiting for signal...")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", err, nil)
	}
	log.Info("Server exiting")
}

// migrate runs goose migrations
func migrate(db *sql.DB, log *logger.Logger) error {
	migrationsDir := "./migrations"

	// Try to run migrations up
	err := goose.Up(db, migrationsDir)
	if err != nil {
		// Check if error is "no next version" (already up-to-date)
		if err.Error() == "no next version found" {
			log.Info("Database already up-to-date, no migrations to run")
			return nil
		}
		return err
	}

	log.Info("Migrations completed successfully")
	return nil
}
