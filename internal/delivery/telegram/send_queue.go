package telegram

import (
	"github.com/m4xvel/monetych_bot/internal/logger"
)

const copyMessageQueueSize = 500

type sendJob struct {
	op     string
	fn     func() error
	fields []any
}

type sendQueue struct {
	jobs chan sendJob
}

func newSendQueue(buffer int) *sendQueue {
	q := &sendQueue{
		jobs: make(chan sendJob, buffer),
	}
	go q.run()
	return q
}

func (q *sendQueue) enqueue(job sendJob) {
	q.jobs <- job
}

func (q *sendQueue) run() {
	for job := range q.jobs {
		if err := retryOnRateLimitForever(job.op, job.fn, job.fields...); err != nil {
			wrapped := wrapTelegramErr(job.op, err)
			keyvals := []any{"op", job.op}
			if len(job.fields) > 0 {
				keyvals = append(keyvals, job.fields...)
			}
			keyvals = append(keyvals, "err", wrapped)
			logger.Log.Errorw("failed to process telegram send job", keyvals...)
		}
	}
}
