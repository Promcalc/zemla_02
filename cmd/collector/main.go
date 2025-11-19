// Package main запускает сервис сбора данных о земельных лотах.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Promcalc/zemla_02/internal/config"
	"github.com/Promcalc/zemla_02/internal/db"
	"github.com/Promcalc/zemla_02/internal/nspd"
	"github.com/Promcalc/zemla_02/internal/rss"
	"github.com/Promcalc/zemla_02/internal/scheduler"
	"github.com/Promcalc/zemla_02/internal/torgi"
	"github.com/Promcalc/zemla_02/internal/version"
)

var (
	showVersion bool
	configPath  string
	verbose     bool
)

func init() {
	flag.BoolVar(&showVersion, "version", false, "Показать версию и выйти")
	flag.StringVar(&configPath, "config", "config/config.yaml", "Путь к конфигурации")
	flag.BoolVar(&verbose, "verbose", false, "Подробное логирование")
}

func main() {
	flag.Parse()

	// Показ версии
	if showVersion {
		fmt.Println(version.String())
		os.Exit(0)
	}

	// Логгер
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	ver := version.Get()
	logger = logger.With(
		"version", ver.Version,
		"commit", ver.Commit,
	)

	slog.SetDefault(logger)

	// Загрузка конфигурации
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Ошибка загрузки конфигурации", "path", configPath, "error", err)
		os.Exit(1)
	}

	// Лог запуска
	logger.Info("Запуск сборщика лотов",
		"version", ver.Version,
		"commit", ver.Commit,
		"config", configPath,
		"rss_url", cfg.RSS.URL,
		"db_url", cfg.Database.URL,
		"schedule_interval", cfg.Collector.ScheduleInterval,
		"retry_interval", cfg.Collector.RetryInterval,
	)

	// Контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-stop
		logger.Info("Получен сигнал завершения", "signal", sig.String())
		cancel()
	}()

	// Инициализация репозитория БД
	dbRepo, err := db.NewRepository(ctx, cfg.Database.URL, logger)
	if err != nil {
		logger.Error("Ошибка подключения к БД", "error", err)
		os.Exit(1)
	}
	defer dbRepo.Close()

	// Инициализация клиентов
	rssParser := rss.NewParser(cfg, logger)
	torgiClient := torgi.NewClient(logger)
	nspdClient := nspd.NewClient(logger)

	// Инициализация nspd (однократно)
	if err := nspdClient.Initialize(ctx); err != nil {
		logger.Warn("Не удалось инициализировать nspd.gov.ru", "error", err)
	}

	// Создание заданий
	collectorJob := scheduler.NewCollectorJob(logger, dbRepo, rssParser, torgiClient, nspdClient)
	retryJob := scheduler.NewRetryJob(logger, dbRepo, torgiClient, nspdClient)

	// Создание и запуск планировщика
	sched := scheduler.New(scheduler.Config{
		CollectorInterval: cfg.Collector.ScheduleInterval,
		RetryInterval:     cfg.Collector.RetryInterval,
	}, logger)

	sched.AddCollectorJob(collectorJob)
	sched.AddRetryJob(retryJob)
	sched.Start()
	defer func() {
		_ = sched.Stop(ctx)
	}()

	logger.Info("Сборщик запущен и ожидает сигнал завершения...")
	<-ctx.Done()

	logger.Info("Завершение работы...")
}
