package job

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) FetchAndLockPendingJob(
	ctx context.Context,
	workerID string,
	leaseDuration time.Duration,
) (*Job, error) {

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT
			id,
			type,
			payload,
			status,
			attempts,
			max_attempts,
			created_at,
			updated_at
		FROM portfolios.job_scheduler.jobs
		WHERE status = 'pending'
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED
		LIMIT 1;
`

	var job Job
	err = tx.QueryRowContext(ctx, query).Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // tidak ada job
	}
	if err != nil {
		return nil, err
	}

	leaseUntil := time.Now().Add(leaseDuration)

	updateQuery := `
		UPDATE portfolios.job_scheduler.jobs
		SET
			status = 'running',
			locked_by = $1,
			locked_at = now(),
			lease_until = $2,
			updated_at = now()
		WHERE id = $3;
`

	_, err = tx.ExecContext(
		ctx,
		updateQuery,
		workerID,
		leaseUntil,
		job.ID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	job.Status = StatusRunning
	job.LockedBy = &workerID
	job.LeaseUntil = &leaseUntil

	return &job, nil
}
