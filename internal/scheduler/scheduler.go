package scheduler

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
	cron              *cron.Cron
	logger            *slog.Logger
	collectorInterval time.Duration
	retryInterval     time.Duration
}

func New(cfg Config, logger *slog.Logger) *Scheduler {
	c := cron.New()
	return &Scheduler{
		cron:              c,
		logger:            logger.With("component", "scheduler"),
		collectorInterval: cfg.CollectorInterval,
		retryInterval:     cfg.RetryInterval,
	}
}

func (s *Scheduler) AddCollectorJob(job *CollectorJob) {
	_, err := s.cron.AddFunc("@every "+s.collectorInterval.String(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		job.Run(ctx)
	})
	if err != nil {
		s.logger.Error("Не удалось добавить задание сбора", "error", err)
	}
}

func (s *Scheduler) AddRetryJob(job *RetryJob) {
	_, err := s.cron.AddFunc("@every "+s.retryInterval.String(), func() {
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
	return nil
}
