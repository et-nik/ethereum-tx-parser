package txparser

import (
	"context"
	"log"
	"time"
)

type worker struct {
	fn func(context.Context) error
}

func newWorker(fn func(context.Context) error) *worker {
	return &worker{fn: fn}
}

func (w *worker) Run(ctx context.Context, period time.Duration) {
	ticker := time.NewTicker(period)

	w.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *worker) tick(ctx context.Context) {
	err := w.fn(ctx)
	if err != nil {
		log.Print(err)
	}
}
