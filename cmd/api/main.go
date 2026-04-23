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

    "loviary.app/backend/internal/application/auth"
    "loviary.app/backend/internal/application/couples"
    "loviary.app/backend/internal/application/home"
    "loviary.app/backend/internal/application/moods"
    "loviary.app/backend/internal/application/reminders"
    "loviary.app/backend/internal/application/streaks"
    "loviary.app/backend/internal/application/users"
    "loviary.app/backend/internal/application/memories"
    "loviary.app/backend/internal/infrastructure/jwt"
    postgres "loviary.app/backend/internal/infrastructure/persistence/postgres"
    "loviary.app/backend/internal/interfaces/http/handlers"
    "loviary.app/backend/internal/interfaces/http/middleware"
    "loviary.app/backend/pkg/config"
    "loviary.app/backend/pkg/db"
    "loviary.app/backend/pkg/logger"
)

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
    if err := migrate(dbConn.DB.DB); err != nil {
        log.Fatal("Failed to run migrations", err)
    }

    // Initialize repositories
    userRepo := postgres.NewUserRepository(dbConn.DB)
    coupleRepo := postgres.NewCoupleRepository(dbConn.DB)
    tokenRepo := postgres.NewRefreshTokenRepository(dbConn.DB)
    moodRepo := postgres.NewMoodRepository(dbConn.DB.DB)
    streakRepo := postgres.NewStreakRepository(dbConn.DB.DB)
    reminderRepo := postgres.NewReminderRepository(dbConn.DB.DB)
    memoryRepo := postgres.NewMemoryRepository(dbConn.DB.DB)

    // Initialize JWT manager
    jwtManager := jwt.NewManager(
        cfg.JWT.Secret,
        cfg.JWT.AccessTokenTTL,
        cfg.JWT.RefreshTokenTTL,
        cfg.JWT.Issuer,
        cfg.JWT.Audience,
    )

    // Initialize services
    userService := users.NewService(userRepo)
    coupleService := couples.NewService(coupleRepo)
    authService := auth.NewService(
        userRepo,
        tokenRepo,
        jwtManager,
        cfg.JWT.AccessTokenTTL,
        cfg.JWT.RefreshTokenTTL,
    )
    moodService := moods.NewService(moodRepo)
    streakService := streaks.NewService(streakRepo)
    reminderService := reminders.NewService(reminderRepo)
    memoryService := memories.NewService(memoryRepo)
    homeService := home.NewService(
        coupleService,
        moodService,
        streakService,
    )

    // Initialize handlers
    userHandler := handlers.NewUserHandler(userService)
    authHandler := handlers.NewAuthHandler(authService)
    coupleHandler := handlers.NewCoupleHandler(coupleService, userService)
    moodHandler := handlers.NewMoodHandler(moodService)
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

    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "healthy",
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
            auth.POST("/login", authHandler.Login)
            auth.POST("/refresh", authHandler.Refresh)
            auth.POST("/logout", authHandler.Logout)
            auth.POST("/logout-all", middleware.AuthMiddleware(jwtManager), authHandler.LogoutAll)
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
            home.GET("/summary", middleware.AuthMiddleware(jwtManager), homeHandler.GetSummary)
        }

        // Storage routes
        storage := v1.Group("/storage")
        {
            storage.POST("/presign", middleware.AuthMiddleware(jwtManager), storageHandler.Presign)
        }
    }

    // Start server
    addr := fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port)
    log.Info("Starting server", map[string]interface{}{"addr": addr})

    srv := &http.Server{
        Addr:    addr,
        Handler: router,
    }

    // Graceful shutdown
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Server error", err)
        }
    }()

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
func migrate(db *sql.DB) error {
    migrationsDir := "./migrations"
    return goose.Up(db, migrationsDir)
}
