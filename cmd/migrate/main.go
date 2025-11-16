// Package main реализует утилиту управления миграциями.
// Поддерживает те же параметры конфигурации, что и основной сервис:
//   - читает db_url из YAML-конфига
//   - поддерживает переопределение через переменную окружения DB_URL
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Promcalc/zemla_02/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // драйвер для PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"       // источник: файлы на диске
)

var (
	action         string
	configPath     string
	migrationsPath string
	version        uint
)

func init() {
	flag.StringVar(&action, "action", "up", "Действие: up, down, force")
	flag.StringVar(&configPath, "config", "config/config.yaml", "Путь к YAML-конфигурации")
	flag.StringVar(&migrationsPath, "migrations", "./migrations", "Путь к папке с миграциями")
	flag.UintVar(&version, "version", 0, "Целевая версия (для down/force)")
}

// Основная логика:
// 1. Загружает конфиг (db_url из config.yaml)
// 2. Поддерживает переопределение DB_URL из окружения (как в основном сервисе)
// 3. Автоматически строит source URL для миграций (file://...)
// 4. Выполняет запрошенное действие
func main() {
	flag.Parse()

	// Загружаем конфигурацию так же, как в основном сервисе
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	// Получаем URL БД: из конфига или из окружения (как в основном коде)
	dbURL := cfg.Database.URL
	if envURL := os.Getenv("DB_URL"); envURL != "" {
		dbURL = envURL
	}

	// Формируем URL источника миграций (file://...)
	absMigPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		log.Fatalf("Не удалось получить абсолютный путь к миграциям: %v", err)
	}
	srcURL := "file://" + absMigPath

	// Инициализируем миграционный движок
	m, err := migrate.New(srcURL, dbURL)
	if err != nil {
		log.Fatalf("Ошибка инициализации миграций: %v", err)
	}
	defer m.Close()

	// Выполняем действие
	var currentVersion uint
	switch action {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Ошибка применения миграций: %v", err)
		}
		currentVersion, _, _ = m.Version()
		fmt.Printf("✅ Миграции применены. Текущая версия: %d\n", currentVersion)

	case "down":
		if version > 0 {
			if err := m.Migrate(version); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Ошибка отката до версии %d: %v", version, err)
			}
		} else {
			if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Ошибка отката одной миграции: %v", err)
			}
		}
		currentVersion, _, _ = m.Version()
		fmt.Printf("↩️  Откат выполнен. Текущая версия: %d\n", currentVersion)

	case "force":
		if err := m.Force(int(version)); err != nil {
			log.Fatalf("Ошибка принудительной установки версии %d: %v", version, err)
		}
		fmt.Printf("⚠️  Версия миграций установлена вручную: %d\n", version)

	default:
		log.Fatalf("Неизвестное действие: %s (допустимо: up, down, force)", action)
	}
}
