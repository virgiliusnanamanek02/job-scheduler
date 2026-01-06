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

func (j *Job) CanRetry() bool {
	return j.Attempts < j.MaxAttempts
}

func (j Job) MarkRunning(workerID string, leaseUntil time.Time) {
	now := time.Now()
	j.Status = StatusRunning
	j.LockedBy = &workerID
	j.LockedAt = &now
	j.LeaseUntil = &leaseUntil
	j.UpdatedAt = now
}

func (j *Job) MarkFailed(err error) {
	now := time.Now()
	msg := err.Error()

	j.Status = StatusFailed
	j.Attempts++
	j.LastError = &msg
	j.UpdatedAt = now
}

func (j *Job) MarkSucceeded() {
	now := time.Now()
	j.Status = StatusSucceeded
	j.UpdatedAt = now
}
