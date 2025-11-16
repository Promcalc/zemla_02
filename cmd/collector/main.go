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

	"github.com/Promcalc/zemla_02/internal/db"
	"github.com/Promcalc/zemla_02/internal/rss"
	"github.com/Promcalc/zemla_02/internal/scheduler"
	"github.com/Promcalc/zemla_02/internal/version"
)

var (
	showVersion bool
	configPath  string
)

func init() {
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.StringVar(&configPath, "config", "config/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	// Показать версию и выйти, если запрошено
	if showVersion {
		fmt.Println(version.String())
		os.Exit(0)
	}

	// Инициализация логгера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Логирование версии при запуске
	logger.Info("Starting lot collector",
		"version", version.Get().Version,
		"commit", version.Get().Commit,
		"build_date", version.Get().BuildDate,
		"go_version", version.Get().GoVersion)

	// Загрузка конфигурации
	// ... ваш код загрузки конфигурации ...

	// Инициализация базы данных
	// ... ваш код инициализации БД ...

	// Создание контекста с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		logger.Info("Shutting down gracefully...")
		cancel()
	}()

	// Запуск планировщика
	scheduler := scheduler.NewCollectorScheduler(
		ctx,
		logger,
		&rss.Collector{},
		&db.Repository{},
		time.Minute*15, // интервал по умолчанию
		time.Hour,      // интервал повторной обработки
	)

	// Запуск основного цикла
	if err := scheduler.Run(ctx); err != nil {
		logger.Error("Scheduler failed", "error", err)
		os.Exit(1)
	}
}
