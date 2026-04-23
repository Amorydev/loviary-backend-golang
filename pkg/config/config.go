package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig
	Database DBConfig
	JWT      JWTConfig
	CORS    CORSConfig
	Rate    RateLimitConfig
	Log     LogConfig
	Storage StorageConfig
	FCM     FCMConfig
	Features FeatureFlags
}

type AppConfig struct {
	Name     string
	Env      string
	Debug    bool
	Port     string
	Host     string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	Timezone string
}

type JWTConfig struct {
	Secret              string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	Issuer              string
	Audience            string
}

type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

type RateLimitConfig struct {
	Enabled  bool
	Requests int
	Window   time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

type StorageConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	Region     string
}

type FCMConfig struct {
	ServerKey string
	ProjectID string
}

type FeatureFlags struct {
	MoodTracking     bool
	Memories         bool
	DailySparks      bool
	Promises         bool
	LoveGoals        bool
	TimeCapsules     bool
}

// Load loads configuration from environment variables
func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("app.port", "8080")
	viper.SetDefault("app.host", "0.0.0.0")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.access_token_ttl", "15m")
	viper.SetDefault("jwt.refresh_token_ttl", "720h")
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests", 100)
	viper.SetDefault("rate_limit.window", "1m")

	// Read config file if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Warning: config file not found, using env vars: %v", err)
		}
	}

	// Manually set values from environment variables
	// This ensures env vars are read correctly in Docker
	if host := os.Getenv("DATABASE_HOST"); host != "" {
		viper.Set("database.host", host)
	}
	if port := os.Getenv("DATABASE_PORT"); port != "" {
		viper.Set("database.port", port)
	}
	if user := os.Getenv("DATABASE_USER"); user != "" {
		viper.Set("database.user", user)
	}
	if password := os.Getenv("DATABASE_PASSWORD"); password != "" {
		viper.Set("database.password", password)
	}
	if name := os.Getenv("DATABASE_NAME"); name != "" {
		viper.Set("database.name", name)
	}
	if sslmode := os.Getenv("DATABASE_SSLMODE"); sslmode != "" {
		viper.Set("database.sslmode", sslmode)
	}
	if timezone := os.Getenv("DATABASE_TIMEZONE"); timezone != "" {
		viper.Set("database.timezone", timezone)
	}
	if appPort := os.Getenv("APP_PORT"); appPort != "" {
		viper.Set("app.port", appPort)
	}
	if appHost := os.Getenv("APP_HOST"); appHost != "" {
		viper.Set("app.host", appHost)
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		viper.Set("jwt.secret", jwtSecret)
	}
	if jwtAccessTTL := os.Getenv("JWT_ACCESS_TOKEN_TTL"); jwtAccessTTL != "" {
		viper.Set("jwt.access_token_ttl", jwtAccessTTL)
	}
	if jwtRefreshTTL := os.Getenv("JWT_REFRESH_TOKEN_TTL"); jwtRefreshTTL != "" {
		viper.Set("jwt.refresh_token_ttl", jwtRefreshTTL)
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		viper.Set("log.level", logLevel)
	}
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		viper.Set("log.format", logFormat)
	}

	// Debug: log database config
	log.Printf("DB config - Host: '%s', Port: '%s', User: '%s', DB: '%s', SSLMode: '%s', Timezone: '%s'",
		viper.GetString("database.host"),
		viper.GetString("database.port"),
		viper.GetString("database.user"),
		viper.GetString("database.name"),
		viper.GetString("database.sslmode"),
		viper.GetString("database.timezone"))

	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		panic("Failed to unmarshal config: " + err.Error())
	}

	return &c
}

// MustLoad loads config or panics if fails
func MustLoad() *Config {
	cfg := Load()
	if cfg.App.Port == "" {
		cfg.App.Port = "8080"
	}
	if cfg.JWT.AccessTokenTTL == 0 {
		cfg.JWT.AccessTokenTTL = 15 * time.Minute
	}
	if cfg.JWT.RefreshTokenTTL == 0 {
		cfg.JWT.RefreshTokenTTL = 720 * time.Hour
	}
	return cfg
}
