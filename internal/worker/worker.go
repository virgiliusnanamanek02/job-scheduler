package worker

import (
	"context"
	"github.com/google/uuid"
	"github.com/virgiliusnanamanek02/job-scheduler/internal/job"
	"log"
	"time"
)

type Worker struct {
	id            string
	repo          *job.Repository
	leaseDuration time.Duration
	pollInterval  time.Duration
}

func New(repo *job.Repository, leaseDuration time.Duration, pollInternal time.Duration) *Worker {
	return &Worker{
		id:            uuid.NewString(),
		repo:          repo,
		leaseDuration: leaseDuration,
		pollInterval:  pollInternal,
	}
}

func (w *Worker) Run(ctx context.Context) {
	log.Printf("[worker:%s] started", w.id)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		for {
			select {
			case <-ctx.Done():
				log.Printf("[worker:%s] shutting down", w.id)
				return

			case <-ticker.C:
				w.processOnce(ctx)
			}
		}
	}
}

func (w *Worker) processOnce(ctx context.Context) {
	jobItem, err := w.repo.FetchAndLockPendingJob(
		ctx,
		w.id,
		w.leaseDuration,
	)
	if err != nil {
		log.Printf("[worker:%s] error fetching job: %v", w.id, err)
		return
	}

	if jobItem == nil {
		// tidak ada job, normal
		return
	}

	log.Printf(
		"[worker:%s] processing job %s (type=%s)",
		w.id,
		jobItem.ID,
		jobItem.Type,
	)

	// ---- simulasi kerja ----
	err = w.handleJob(jobItem)

	if err != nil {
		log.Printf(
			"[worker:%s] job %s failed: %v",
			w.id,
			jobItem.ID,
			err,
		)
		// TODO: update status failed + retry
		return
	}

	log.Printf(
		"[worker:%s] job %s completed",
		w.id,
		jobItem.ID,
	)
	// TODO: update status succeeded
}

func (w *Worker) handleJob(j *job.Job) error {
	// simulasi kerja
	time.Sleep(2 * time.Second)

	// TODO:
	// - switch by job.Type
	// - decode payload
	// - execute real logic

	return nil
}
