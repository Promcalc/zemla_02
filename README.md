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

# Ошибки в логе при работе

{"time":"2025-11-16T22:30:10.873118666Z","level":"INFO","msg":"Запрос деталей лота","service":"torgi","lot_id":"24000029350000000060_21","rss_link":"https://torgi.gov.ru/new/public/lots/lot/24000029350000000060_21"}
{"time":"2025-11-16T22:30:11.419346875Z","level":"DEBUG","msg":"Успешно получены детали лота","service":"torgi","lot_id":"24000029350000000060_21","keys":55}
{"time":"2025-11-16T22:30:11.419486284Z","level":"INFO","msg":"Запрос данных по кадастровому номеру","service":"nspd","cn":"56:03:0000000:13"}
{"time":"2025-11-16T22:30:12.018513781Z","level":"DEBUG","msg":"Успешно получены данные geoportal","service":"nspd","cn":"56:03:0000000:13","features":0}
{"time":"2025-11-16T22:30:12.019768963Z","level":"DEBUG","msg":"Лот с таким GUID уже существует","guid":"https://torgi.gov.ru/new/public/lots/lot/24000029350000000060_21"}
{"time":"2025-11-16T22:30:12.02018579Z","level":"ERROR","msg":"Ошибка сохранения лота","job":"collector","guid":"https://torgi.gov.ru/new/public/lots/lot/24000029350000000060_21","error":"ошибка сохранения лота: не удалось найти существующий лот: ERROR: current transaction is aborted, commands ignored until end of transaction block (SQLSTATE 25P02)"}
{"time":"2025-11-16T22:30:12.020269796Z","level":"INFO","msg":"Запрос деталей лота","service":"torgi","lot_id":"22000036770000000134_4","rss_link":"https://torgi.gov.ru/new/public/lots/lot/22000036770000000134_4"}
{"time":"2025-11-16T22:30:12.582798537Z","level":"DEBUG","msg":"Успешно получены детали лота","service":"torgi","lot_id":"22000036770000000134_4","keys":53}
{"time":"2025-11-16T22:30:12.582907245Z","level":"INFO","msg":"Запрос данных по кадастровому номеру","service":"nspd","cn":"65:11:0000020:481"}
{"time":"2025-11-16T22:30:13.201356529Z","level":"DEBUG","msg":"Успешно получены данные geoportal","service":"nspd","cn":"65:11:0000020:481","features":0}
{"time":"2025-11-16T22:30:13.202420398Z","level":"DEBUG","msg":"Лот с таким GUID уже существует","guid":"https://torgi.gov.ru/new/public/lots/lot/22000036770000000134_4"}
{"time":"2025-11-16T22:30:13.202689516Z","level":"ERROR","msg":"Ошибка сохранения лота","job":"collector","guid":"https://torgi.gov.ru/new/public/lots/lot/22000036770000000134_4","error":"ошибка сохранения лота: не удалось найти существующий лот: ERROR: current transaction is aborted, commands ignored until end of transaction block (SQLSTATE 25P02)"}
{"time":"2025-11-16T22:30:13.202775021Z","level":"INFO","msg":"Запрос деталей лота","service":"torgi","lot_id":"22000036770000000134_2","rss_link":"https://torgi.gov.ru/new/public/lots/lot/22000036770000000134_2"}
{"time":"2025-11-16T22:30:13.751601071Z","level":"DEBUG","msg":"Успешно получены детали лота","service":"torgi","lot_id":"22000036770000000134_2","keys":53}
{"time":"2025-11-16T22:30:13.751626672Z","level":"INFO","msg":"Запрос данных по кадастровому номеру","service":"nspd","cn":"65:11:0000020:479"}
{"time":"2025-11-16T22:30:16.257034269Z","level":"DEBUG","msg":"Успешно получены данные geoportal","service":"nspd","cn":"65:11:0000020:479","features":0}
{"time":"2025-11-16T22:30:16.258482564Z","level":"DEBUG","msg":"Лот с таким GUID уже существует","guid":"https://torgi.gov.ru/new/public/lots/lot/22000036770000000134_2"}
{"time":"2025-11-16T22:30:16.258900091Z","level":"ERROR","msg":"Ошибка сохранения лота","job":"collector","guid":"https://torgi.gov.ru/new/public/lots/lot/22000036770000000134_2","error":"ошибка сохранения лота: не удалось найти существующий лот: ERROR: current transaction is aborted, commands ignored until end of transaction block (SQLSTATE 25P02)"}
{"time":"2025-11-16T22:30:16.258920992Z","level":"INFO","msg":"Задание сбора данных завершено","job":"collector"}
^C{"time":"2025-11-16T22:35:39.72350187Z","level":"INFO","msg":"Получен сигнал завершения","signal":"interrupt"}
{"time":"2025-11-16T22:35:39.724090003Z","level":"INFO","msg":"Завершение работы..."}
{"time":"2025-11-16T22:35:39.724102304Z","level":"INFO","msg":"Остановка планировщика","component":"scheduler"}