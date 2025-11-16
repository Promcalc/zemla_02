# Lot Collector Service

Сервис мониторинга земельных лотов с отображением на Яндекс.Картах. Собирает данные из RSS-ленты аукционов, обогащает их информацией из внешних API и сохраняет в базу данных с возможностью повторной обработки ошибок.

## 🚀 Особенности

- ✅ Регулярный сбор данных из RSS-ленты torgi.gov.ru
- ✅ Обогащение данных информацией из API torgi.gov.ru и nspd.gov.ru
- ✅ Поддержка повторной обработки ошибок с экспоненциальной задержкой
- ✅ Пространственные запросы с использованием PostGIS
- ✅ Агрегация данных в зависимости от масштаба карты
- ✅ Отказоустойчивая архитектура с health checks
- ✅ Контейнеризация через Docker
- ✅ CI/CD через GitHub Actions

## 🛠 Технологии

- **Язык**: Go 1.25
- **База данных**: PostgreSQL 16 с PostGIS 3.4
- **Веб-фреймворк**: Gin
- **Планировщик**: robfig/cron
- **HTTP-клиент**: resty
- **Парсинг RSS**: gofeed
- **Пространственные данные**: orb, twpb/geometry
- **Миграции**: golang-migrate

## 📁 Структура проекта

lot-collector/
├── cmd/ # Точки входа приложений
│ ├── collector/ # Сборщик данных
│ └── web/ # Веб-сервер
├── internal/ # Внутренние пакеты
│ ├── rss/ # Работа с RSS
│ ├── torgi/ # API torgi.gov.ru
│ ├── nspd/ # API nspd.gov.ru
│ ├── db/ # Работа с базой данных
│ ├── scheduler/ # Планировщик задач
│ └── api/ # HTTP API
├── web/ # Веб-интерфейс
│ ├── static/ # Статические файлы
│ └── templates/ # Шаблоны
├── migrations/ # Миграции базы данных
├── config/ # Конфигурация
├── local-dev/ # Локальная разработка
├── .github/workflows/ # GitHub Actions
├── Dockerfile # Docker образ
├── docker-compose.* # Docker Compose файлы
├── Makefile # Команды для разработки
├── go.mod # Зависимости Go
└── README.md # Документация



## 🏷️ Версионирование

Проект использует семантическое версионирование (SemVer). Версии определяются через Git-теги в форматах:

- `v1.2.3` (рекомендуемый формат)
- `1.2.3` (также поддерживается)

### Автоматическое управление версиями

При пуше в `main` ветку:
- Сборка с тегом `latest` и `<short-commit-sha>`
- Версия `dev` в бинарных файлах

При создании Git-тега:
- Сборка с тегами `<version>` и `<short-commit-sha>`
- Автоматическое обновление тега `latest`
- Версия из тега в бинарных файлах

### Просмотр версии

```bash
# В контейнере
docker run --rm your-dockerhub-username/lot-collector:latest --version

# Локально
make version
bin/collector --version
```

# Запуск отладчика:

В контейнере делаем так:
```
dlv debug --headless --listen=:2345 --api-version=2 --log --log-output=debugger --accept-multiclient --continue=false ./cmd/collector
```

В vscode  - Start debugging F5

Запуск миграций:

```
go run ./cmd/migrate --config config/config.yaml --action up
```
