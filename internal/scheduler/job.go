// Package scheduler содержит задания для регулярного выполнения.
package scheduler

import (
	"context"
	"log/slog"

	"github.com/Promcalc/zemla_02/internal/db"
	"github.com/Promcalc/zemla_02/internal/nspd"
	"github.com/Promcalc/zemla_02/internal/parser"
	"github.com/Promcalc/zemla_02/internal/rss"
	"github.com/Promcalc/zemla_02/internal/torgi"
)

// CollectorJob выполняет полный цикл сбора данных.
type CollectorJob struct {
	logger      *slog.Logger
	dbRepo      *db.Repository
	rssParser   *rss.Parser
	torgiClient *torgi.Client
	nspdClient  *nspd.Client
}

func NewCollectorJob(
	logger *slog.Logger,
	dbRepo *db.Repository,
	rssParser *rss.Parser,
	torgiClient *torgi.Client,
	nspdClient *nspd.Client,
) *CollectorJob {
	return &CollectorJob{
		logger:      logger,
		dbRepo:      dbRepo,
		rssParser:   rssParser,
		torgiClient: torgiClient,
		nspdClient:  nspdClient,
	}
}

// Run выполняет сбор новых лотов из RSS и обогащение данными.
func (j *CollectorJob) Run(ctx context.Context) {
	logger := j.logger.With("job", "collector")
	logger.Info("Запуск задания сбора данных")

	// Получаем последнюю дату из БД
	lastPubDate, err := j.dbRepo.GetLastPubDate(ctx)
	if err != nil {
		logger.Error("Не удалось получить last_pub_date", "error", err)
		return
	}

	// Получаем только новые лоты
	lots, err := j.rssParser.FetchAndParseSince(ctx, lastPubDate)
	if err != nil {
		logger.Error("Ошибка при получении RSS", "error", err)
		return
	}

	// 1. Получаем RSS
	// lots, err := j.rssParser.FetchAndParse(ctx)
	// if err != nil {
	// 	logger.Error("Ошибка при получении RSS", "error", err)
	// 	return
	// }

	if len(lots) == 0 {
		logger.Info("Новые лоты не найдены")
		return
	}

	logger.Info("Найдено новых лотов", "count", len(lots))

	// 2. Инициализируем nspd (если ещё не инициализирован)
	if err := j.nspdClient.Initialize(ctx); err != nil {
		logger.Warn("Не удалось инициализировать nspd", "error", err)
		// Продолжаем без nspd — ошибки будут повторно обработаны позже
	}

	// 3. Обрабатываем каждый лот
	for _, rssLot := range lots {
		select {
		case <-ctx.Done():
			logger.Info("Задание прервано по контексту")
			return
		default:
		}

		// Извлекаем кадастровый номер и дату аукциона
		extracted := parser.ExtractFromLot(rssLot)

		// Запрос к torgi.gov.ru
		var lotInfo map[string]interface{}
		var torgiErr error
		if rssLot.Link != "" {
			lotInfo, torgiErr = j.torgiClient.GetLotInfo(ctx, rssLot.Link)
		}

		// Запрос к nspd.gov.ru (только если есть кадастровый номер)
		var nspdResp *nspd.GeoportalResponse
		var nspdErr error
		if extracted.CadastralNumber != "" {
			nspdResp, nspdErr = j.nspdClient.SearchByCadastralNumber(ctx, extracted.CadastralNumber)
		}

		// Сохраняем всё атомарно
		_, err := j.dbRepo.SaveLotWithExternalData(
			ctx, rssLot, extracted,
			lotInfo, nspdResp,
			torgiErr, nspdErr,
		)
		if err != nil {
			logger.Error("Ошибка сохранения лота", "guid", rssLot.GUID, "error", err)
		} else {
			logger.Debug("Лот успешно сохранён", "guid", rssLot.GUID)
		}
	}

	logger.Info("Задание сбора данных завершено")
}

// RetryJob выполняет повторную обработку ошибок.
type RetryJob struct {
	logger      *slog.Logger
	dbRepo      *db.Repository
	torgiClient *torgi.Client
	nspdClient  *nspd.Client
}

func NewRetryJob(
	logger *slog.Logger,
	dbRepo *db.Repository,
	torgiClient *torgi.Client,
	nspdClient *nspd.Client,
) *RetryJob {
	return &RetryJob{
		logger:      logger,
		dbRepo:      dbRepo,
		torgiClient: torgiClient,
		nspdClient:  nspdClient,
	}
}

// Run повторно запрашивает данные для записей с ошибками.
func (j *RetryJob) Run(ctx context.Context) {
	logger := j.logger.With("job", "retry")
	logger.Info("Запуск задания повторной обработки ошибок")

	// TODO: реализуйте метод FindLotWithErrors в db.Repository
	// Пока что заглушка — можно реализовать позже
	// Для первого этапа достаточно лога

	logger.Info("Повторная обработка ошибок не реализована (заглушка)")
	// В будущем: dbRepo.FindLotsWithErrors() → повторить запросы → обновить external_data
}
