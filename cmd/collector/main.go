package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Promcalc/zemla_02/internal/config"
	"github.com/Promcalc/zemla_02/internal/version"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	showVersion bool
	configPath  string
	verbose     bool
)

func init() {
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.StringVar(&configPath, "config", "config/config.yaml", "Path to configuration file")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
}

func main() {
	flag.Parse()

	// Показать версию и выйти, если запрошено
	if showVersion {
		fmt.Println(version.String())
		os.Exit(0)
	}

	// Настройка логера
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	// Загрузка конфигурации
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err, "config_path", configPath)
		os.Exit(1)
	}

	// Логирование запуска сервиса
	logger.Info("Starting lot collector service",
		"version", version.Get().Version,
		"commit", version.Get().Commit,
		"build_date", version.Get().BuildDate,
		"go_version", version.Get().GoVersion,
		"config_path", configPath,
		"rss_url", cfg.RSS.URL,
		"db_url", cfg.Database.URL,
		"schedule_interval", cfg.Collector.ScheduleInterval,
		"retry_interval", cfg.Collector.RetryInterval,
	)

	// Логирование конфигурации с флагом debug
	if verbose {
		logger.Debug("Full configuration",
			"config", fmt.Sprintf("%+v", cfg),
		)
	}

	// Создание контекста с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-stop
		logger.Info("Received shutdown signal", "signal", sig.String())
		cancel()
	}()

	// Имитация работы сервиса
	logger.Info("Service started successfully. Waiting for shutdown signal...")
	<-ctx.Done()

	// Graceful shutdown
	logger.Info("Starting graceful shutdown...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Здесь будет логика graceful shutdown (закрытие соединений с БД и т.д.)
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout reached, forcing exit")
	case <-time.After(2 * time.Second):
		logger.Info("Graceful shutdown completed")
	}

	logger.Info("Service stopped")
}
