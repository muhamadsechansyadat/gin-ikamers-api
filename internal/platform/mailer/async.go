package mailer

import (
	"context"
	"log/slog"
	"time"
)

type AsyncMailer struct {
	inner   Mailer
	logger  *slog.Logger
	timeout time.Duration
}

func NewAsync(inner Mailer, logger *slog.Logger) *AsyncMailer {
	return &AsyncMailer{
		inner:   inner,
		logger:  logger,
		timeout: 30 * time.Second,
	}
}

func (a *AsyncMailer) Send(_ context.Context, to, subject, htmlBody string) error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.logger.Error("mailer panic recovered",
					"panic", r, "to", to, "subject", subject)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
		defer cancel()

		start := time.Now()
		if err := a.inner.Send(ctx, to, subject, htmlBody); err != nil {
			a.logger.Error("mailer send failed",
				"error", err, "to", to, "subject", subject,
				"duration_ms", time.Since(start).Milliseconds())
			return
		}
		a.logger.Info("mailer sent",
			"to", to, "subject", subject,
			"duration_ms", time.Since(start).Milliseconds())
	}()
	return nil
}
