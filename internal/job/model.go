package job

import (
	"github.com/google/uuid"
	"github.com/virgiliusnanamanek02/job-scheduler/constanta"
	"time"
)

type Status string

const StatusPending Status = constanta.StatusPending
const StatusRunning Status = constanta.StatusRunning
const StatusSucceeded Status = constanta.StatusSucceed
const StatusFailed Status = constanta.StatusFailed

type Job struct {
	ID          uuid.UUID
	Type        string
	Payload     []byte
	Status      Status
	LockedBy    *string
	LockedAt    *time.Time
	LeaseUntil  *time.Time
	Attempts    int
	MaxAttempts int
	LastError   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (job *Job) CanRetry() bool {
	return job.Attempts < job.MaxAttempts
}

func (job Job) MarkRunning(workerID string, leaseUntil time.Time) {
	now := time.Now()
	job.Status = StatusRunning
	job.LockedBy = &workerID
	job.LockedAt = &now
	job.LeaseUntil = &leaseUntil
	job.UpdatedAt = now
}

func (job *Job) MarkFailed(err error) {
	now := time.Now()
	msg := err.Error()

	job.Status = StatusFailed
	job.Attempts++
	job.LastError = &msg
	job.UpdatedAt = now
}

func (job *Job) MarkSucceeded() {
	now := time.Now()
	job.Status = StatusSucceeded
	job.UpdatedAt = now
}
