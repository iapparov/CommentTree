package config

import (
	"fmt"
	wbfconfig "github.com/wb-go/wbf/config"
	"os"
	"time"
)

type AppConfig struct {
	ServerConfig ServerConfig `mapstructure:"server"`
	LoggerConfig loggerConfig `mapstructure:"logger"`
	RedisConfig  redisConfig  `mapstructure:"redis"`
	DBConfig     dbConfig     `mapstructure:"db_config"`
	RetrysConfig RetrysConfig `mapstructure:"retry_strategy"`
	GinConfig    ginConfig    `mapstructure:"gin"`
}

type RetrysConfig struct {
	Attempts int           `mapstructure:"attempts" default:"3"`
	Delay    time.Duration `mapstructure:"delay" default:"1s"`
	Backoffs float64       `mapstructure:"backoffs" default:"2"`
}

type ginConfig struct {
	Mode string `mapstructure:"mode" default:"debug"`
}

type ServerConfig struct {
	Host string `mapstructure:"host" default:"localhost"`
	Port int    `mapstructure:"port" default:"8080"`
}

type loggerConfig struct {
	Level string `mapstructure:"level" default:"info"`
}

type redisConfig struct {
	Host      string `mapstructure:"host" default:"localhost"`
	Port      int    `mapstructure:"port" default:"6379"`
	Password  string `mapstructure:"password" default:""`
	DB        int    `mapstructure:"db" default:"0"`
	TTL       string `mapstructure:"ttl" default:"30s"`
	CacheSize int    `mapstructure:"cache_size" default:"1000"`
}

type postgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode" default:"disable"`
}

type dbConfig struct {
	Master          postgresConfig   `mapstructure:"postgres"`
	Slaves          []postgresConfig `mapstructure:"slaves"`
	MaxOpenConns    int              `mapstructure:"maxOpenConns"`
	MaxIdleConns    int              `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration    `mapstructure:"connMaxLifetime"`
}

func NewAppConfig() (*AppConfig, error) {
	envFilePath := "./.env"
	appConfigFilePath := "./config/local.yaml"

	cfg := wbfconfig.New()

	// Загрузка .env файлов
	if err := cfg.LoadEnvFiles(envFilePath); err != nil {
		return nil, fmt.Errorf("failed to load env files: %w", err)
	}

	// Включение поддержки переменных окружения
	cfg.EnableEnv("")

	if err := cfg.LoadConfigFiles(appConfigFilePath); err != nil {
		return nil, fmt.Errorf("failed to load config files: %w", err)
	}

	var appCfg AppConfig
	if err := cfg.Unmarshal(&appCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	appCfg.DBConfig.Master.DBName = os.Getenv("POSTGRES_DB")
	appCfg.DBConfig.Master.User = os.Getenv("POSTGRES_USER")
	appCfg.DBConfig.Master.Password = os.Getenv("POSTGRES_PASSWORD")

	appCfg.RedisConfig.Password = os.Getenv("REDIS_PASSWORD")

	return &appCfg, nil
}
