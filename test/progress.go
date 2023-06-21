// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/stv0g/cunicu/pkg/log"
)

type ProgressHandler interface {
	OnStart()
	OnEnd(cntCompleted, cntFailed uint, durElapsed time.Duration)
	OnProgress(cntStarted, cntCompleted, cntFailed uint, durElapsed, durRemaining time.Duration, idsMissing []string)
	OnError(err error)
}

func WithProgress(ctx context.Context, run func(started, completed chan string) error, handler ProgressHandler) error {
	ids := map[string]any{}

	errors := make(chan error)
	started := make(chan string)
	completed := make(chan string)
	done := make(chan any)

	var cntStarted, cntCompleted, cntFailed, cntLast uint

	go func() {
		if err := run(started, completed); err != nil {
			errors <- err
		}
		close(done)
	}()

	handler.OnStart()
	start := time.Now()

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			if cntLast != cntCompleted {
				percentage := float64(cntCompleted) / float64(cntStarted)
				durElapsed := time.Since(start)
				durRemaining := time.Duration(float64(durElapsed) * (1 - percentage))

				handler.OnProgress(cntStarted, cntCompleted, cntFailed, durElapsed, durRemaining, maps.Keys(ids))

				cntLast = cntCompleted
			}
		case id := <-started:
			cntStarted++
			ids[id] = nil

		case id := <-completed:
			cntCompleted++
			delete(ids, id)

		case <-done:
			durElapsed := time.Since(start)
			handler.OnEnd(cntCompleted, cntFailed, durElapsed)

			if cntFailed > 0 {
				return fmt.Errorf("%d runs failed", cntFailed)
			}

			return nil

		case <-ctx.Done():
			return ctx.Err()

		case err := <-errors:
			cntFailed++
			handler.OnError(err)
		}
	}
}

var _ ProgressHandler = (*DefaultProgressHandler)(nil)

type DefaultProgressHandler struct {
	Logger *log.Logger
}

func (ph *DefaultProgressHandler) OnProgress(cntStarted, cntCompleted, cntFailed uint, durElapsed, durRemaining time.Duration, idsMissing []string) {
	fields := []zap.Field{
		zap.Int("percentage", int(100*cntCompleted/cntStarted)),
		zap.Uint("started", cntStarted),
		zap.Uint("completed", cntCompleted),
		zap.Uint("failed", cntFailed),
		zap.Duration("elapsed", durElapsed),
		zap.Duration("remaining", durRemaining),
	}

	if len(idsMissing) < 10 {
		fields = append(fields,
			zap.Strings("missing", idsMissing),
		)
	}

	ph.Logger.Info("Progress", fields...)
}

func (ph *DefaultProgressHandler) OnStart() {
	if ph.Logger == nil {
		ph.Logger = log.Global.Named("progress")
	}

	ph.Logger.Info("Started")
}

func (ph *DefaultProgressHandler) OnEnd(cntCompleted, cntFailed uint, durElapsed time.Duration) {
	ph.Logger.Info("Completed",
		zap.Duration("elapsed", durElapsed),
		zap.Uint("completed", cntCompleted),
		zap.Uint("failed", cntFailed))
}

func (ph *DefaultProgressHandler) OnError(err error) {
	ph.Logger.Error("Failed", zap.Error(err))
}
