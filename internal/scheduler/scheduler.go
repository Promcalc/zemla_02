package scheduler

/*
// После создания клиентов и репозитория:

// Создаём задания
collectorJob := scheduler.NewCollectorJob(logger, dbRepo, rssParser, torgiClient, nspdClient)
retryJob := scheduler.NewRetryJob(logger, dbRepo, torgiClient, nspdClient)

// Создаём планировщик
sched := scheduler.New(scheduler.Config{
	CollectorInterval: cfg.Collector.ScheduleInterval,
	RetryInterval:     cfg.Collector.RetryInterval,
}, logger)

// Регистрируем задания
sched.AddCollectorJob(collectorJob)
sched.AddRetryJob(retryJob)

// Запускаем
sched.Start()
defer sched.Stop(ctx)

// Ожидаем сигнал завершения
<-ctx.Done()
*/

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type Config struct {
	CollectorInterval time.Duration
	RetryInterval     time.Duration
}

type Scheduler struct {
	cron   *cron.Cron
	logger *slog.Logger
}

func New(cfg Config, logger *slog.Logger) *Scheduler {
	c := cron.New()
	return &Scheduler{
		cron:   c,
		logger: logger.With("component", "scheduler"),
	}
}

func (s *Scheduler) AddCollectorJob(job *CollectorJob) {
	_, err := s.cron.AddFunc("@every "+cfg.CollectorInterval.String(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		job.Run(ctx)
	})
	if err != nil {
		s.logger.Error("Не удалось добавить задание сбора", "error", err)
	}
}

func (s *Scheduler) AddRetryJob(job *RetryJob) {
	_, err := s.cron.AddFunc("@every "+cfg.RetryInterval.String(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		job.Run(ctx)
	})
	if err != nil {
		s.logger.Error("Не удалось добавить задание повтора", "error", err)
	}
}

func (s *Scheduler) Start() {
	s.logger.Info("Запуск планировщика")
	s.cron.Start()
}

func (s *Scheduler) Stop(ctx context.Context) error {
	s.logger.Info("Остановка планировщика")
	s.cron.Stop()
	<-s.cron.StopCh()
	return nil
}
