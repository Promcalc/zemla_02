package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"go.yaml.in/yaml/v2"
)

// Config представляет полную конфигурацию приложения
type Config struct {
	RSS       RSSConfig       `yaml:"rss"`
	Database  DatabaseConfig  `yaml:"database"`
	Web       WebConfig       `yaml:"web"`
	Collector CollectorConfig `yaml:"collector"`
	External  ExternalConfig  `yaml:"external_apis"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// RSSConfig содержит настройки RSS-ленты
type RSSConfig struct {
	URL       string        `yaml:"url"`
	Timeout   time.Duration `yaml:"timeout"`
	MaxItems  int           `yaml:"max_items"`
	IgnoreSSL bool          `yaml:"ignore_ssl"`
}

// DatabaseConfig содержит настройки базы данных
type DatabaseConfig struct {
	URL              string        `yaml:"url"`
	MaxConnections   int           `yaml:"max_connections"`
	ConnectTimeout   time.Duration `yaml:"connect_timeout"`
	StatementTimeout time.Duration `yaml:"statement_timeout"`
	IdleTimeout      time.Duration `yaml:"idle_timeout"`
	MaxIdleConns     int           `yaml:"max_idle_connections"`
}

// WebConfig содержит настройки веб-сервера
type WebConfig struct {
	Port         int           `yaml:"port"`
	Host         string        `yaml:"host"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// CollectorConfig содержит настройки сборщика данных
type CollectorConfig struct {
	ScheduleInterval time.Duration `yaml:"schedule_interval"`
	RetryInterval    time.Duration `yaml:"retry_interval"`
	MaxRetries       int           `yaml:"max_retries"`
	Concurrency      int           `yaml:"concurrency"`
	BackoffInitial   time.Duration `yaml:"backoff_initial"`
	BackoffMax       time.Duration `yaml:"backoff_max"`
}

// ExternalConfig содержит настройки внешних API
type ExternalConfig struct {
	Torgi TorgiAPIConfig `yaml:"torgi"`
	NSPD  NSPDAPIConfig  `yaml:"nspd"`
}

// TorgiAPIConfig содержит настройки API torgi.gov.ru
type TorgiAPIConfig struct {
	BaseURL    string        `yaml:"base_url"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
	RetryDelay time.Duration `yaml:"retry_delay"`
	UserAgent  string        `yaml:"user_agent"`
	IgnoreSSL  bool          `yaml:"ignore_ssl"`
}

// NSPDAPIConfig содержит настройки API nspd.gov.ru
type NSPDAPIConfig struct {
	BaseURL string        `yaml:"base_url"`
	MapURL  string        `yaml:"map_url"`
	Timeout time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
	RetryDelay time.Duration `yaml:"retry_delay"`
	UserAgent  string        `yaml:"user_agent"`
	Referer    string        `yaml:"referer"`
	IgnoreSSL  bool          `yaml:"ignore_ssl"`
}

// LoggingConfig содержит настройки логирования
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load загружает конфигурацию из файла
func Load(path string) (*Config, error) {
	// Возвращаем дефолтные значения
	cfg := &Config{
		RSS: RSSConfig{
			URL:       getEnv("RSS_URL", "https://torgi.gov.ru/new/api/public/lotcards/rss?lotStatus=PUBLISHED,APPLICATIONS_SUBMISSION&catCode=2&byFirstVersion=true"),
			Timeout:   getDurationEnv("RSS_TIMEOUT", 30*time.Second),
			MaxItems:  getIntEnv("RSS_MAX_ITEMS", 100),
			IgnoreSSL: getBoolEnv("RSS_IGNORE_SSL", true),
		},
		Database: DatabaseConfig{
			URL:              getEnv("DB_URL", "postgres://postgres:postgres@db:5432/lots?sslmode=disable"),
			MaxConnections:   getIntEnv("DB_MAX_CONNECTIONS", 10),
			ConnectTimeout:   getDurationEnv("DB_CONNECT_TIMEOUT", 5*time.Second),
			StatementTimeout: getDurationEnv("DB_STATEMENT_TIMEOUT", 30*time.Second),
			IdleTimeout:      getDurationEnv("DB_IDLE_TIMEOUT", 30*time.Minute),
			MaxIdleConns:     getIntEnv("DB_MAX_IDLE_CONNECTIONS", 2),
		},
		Web: WebConfig{
			Port:         getIntEnv("WEB_PORT", 8000),
			Host:         getEnv("WEB_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("WEB_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("WEB_WRITE_TIMEOUT", 15*time.Second),
		},
		Collector: CollectorConfig{
			ScheduleInterval: getDurationEnv("COLLECTOR_SCHEDULE_INTERVAL", 15*time.Minute),
			RetryInterval:    getDurationEnv("COLLECTOR_RETRY_INTERVAL", 1*time.Hour),
			MaxRetries:       getIntEnv("COLLECTOR_MAX_RETRIES", 5),
			Concurrency:      getIntEnv("COLLECTOR_CONCURRENCY", 3),
			BackoffInitial:   getDurationEnv("COLLECTOR_BACKOFF_INITIAL", 1*time.Second),
			BackoffMax:       getDurationEnv("COLLECTOR_BACKOFF_MAX", 5*time.Minute),
		},
		External: ExternalConfig{
			Torgi: TorgiAPIConfig{
				BaseURL:    getEnv("TORGI_BASE_URL", "https://torgi.gov.ru/new/api/public/lotcards"),
				Timeout:    getDurationEnv("TORGI_TIMEOUT", 30*time.Second),
				MaxRetries: getIntEnv("TORGI_MAX_RETRIES", 3),
				RetryDelay: getDurationEnv("TORGI_RETRY_DELAY", 2*time.Second),
				UserAgent:  getEnv("TORGI_USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
				IgnoreSSL:  getBoolEnv("TORGI_IGNORE_SSL", true),
			},
			NSPD: NSPDAPIConfig{
				BaseURL:    getEnv("NSPD_BASE_URL", "https://nspd.gov.ru/api/geoportal/v2/search/geoportal"),
				MapURL:     getEnv("NSPD_MAP_URL", "https://nspd.gov.ru/map?thematic=PKK"),
				Timeout:    getDurationEnv("NSPD_TIMEOUT", 45*time.Second),
				MaxRetries: getIntEnv("NSPD_MAX_RETRIES", 3),
				RetryDelay: getDurationEnv("NSPD_RETRY_DELAY", 3*time.Second),
				UserAgent:  getEnv("NSPD_USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
				Referer:    getEnv("NSPD_REFERER", "https://nspd.gov.ru/map"),
				IgnoreSSL:  getBoolEnv("NSPD_IGNORE_SSL", true),
			},
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// 2. Пытаемся прочитать YAML
	data, err := os.ReadFile(path)
	if err != nil {
		// Если файл не найден — используем fallback
		if os.IsNotExist(err) {
			fmt.Printf("⚠️  Конфигурационный файл %s не найден, используются значения по умолчанию\n", path)
			return cfg, nil
		}
		return nil, fmt.Errorf("не удалось прочитать файл конфигурации: %w", err)
	}

	// 3. Парсим YAML поверх fallback-значений
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга config.yaml: %w", err)
	}

	// 4. Переопределяем из окружения (если задано)
	if url := os.Getenv("RSS_URL"); url != "" {
		cfg.RSS.URL = url
	}
	if url := os.Getenv("DB_URL"); url != "" {
		cfg.Database.URL = url
	}

	return cfg, nil

}

// Вспомогательные функции для чтения переменных окружения

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
