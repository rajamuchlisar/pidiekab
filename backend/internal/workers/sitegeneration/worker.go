package sitegeneration

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

type Worker struct {
	jobRepo      *JobRepository
	generator    *Generator
	releaseMgr   *ReleaseManager
	pollInterval time.Duration
	maxAttempts  int
	logger       *slog.Logger
	stopOnce     sync.Once
	doneCh       chan struct{}
}

func NewWorker(
	jobRepo *JobRepository,
	generator *Generator,
	releaseMgr *ReleaseManager,
	pollInterval time.Duration,
	maxAttempts int,
	logger *slog.Logger,
) *Worker {
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Worker{
		jobRepo:      jobRepo,
		generator:    generator,
		releaseMgr:   releaseMgr,
		pollInterval: pollInterval,
		maxAttempts:  maxAttempts,
		logger:       logger,
		doneCh:       make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("site generation worker started",
		"poll_interval", w.pollInterval.String(),
		"max_attempts", w.maxAttempts,
	)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	defer close(w.doneCh)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("site generation worker received shutdown signal")
			return
		case <-ticker.C:
			w.processNextJob(ctx)
		}
	}
}

// ProcessSingleJob processes one pending job immediately (useful for synchronous triggers or tests).
func (w *Worker) ProcessSingleJob(ctx context.Context) (bool, error) {
	job, err := w.jobRepo.ClaimNextJob(ctx)
	if err != nil {
		if errors.Is(err, ErrNoPendingJob) {
			return false, nil
		}
		return false, err
	}

	w.logger.Info("claimed site generation job", "job_id", job.ID, "job_type", job.JobType, "attempt", job.Attempts)

	releaseID, stagingPath, err := w.releaseMgr.CreateStagingDir()
	if err != nil {
		w.logger.Error("failed to create staging directory", "job_id", job.ID, "error", err)
		_ = w.jobRepo.MarkFailed(ctx, job.ID, err.Error(), 10*time.Second, w.maxAttempts)
		return true, err
	}

	buildStart := time.Now()
	if err := w.generator.Generate(ctx, stagingPath); err != nil {
		w.logger.Error("site generation failed", "job_id", job.ID, "error", err)
		_ = w.jobRepo.MarkFailed(ctx, job.ID, err.Error(), 30*time.Second, w.maxAttempts)
		return true, err
	}

	activePath, err := w.releaseMgr.ActivateRelease(releaseID, stagingPath)
	if err != nil {
		w.logger.Error("failed to activate release", "job_id", job.ID, "error", err)
		_ = w.jobRepo.MarkFailed(ctx, job.ID, err.Error(), 30*time.Second, w.maxAttempts)
		return true, err
	}

	if err := w.jobRepo.MarkSucceeded(ctx, job.ID); err != nil {
		w.logger.Error("failed to mark job as succeeded", "job_id", job.ID, "error", err)
		return true, err
	}

	w.logger.Info("site generation release activated successfully",
		"job_id", job.ID,
		"release_id", releaseID,
		"active_path", activePath,
		"duration", time.Since(buildStart).String(),
	)

	return true, nil
}

func (w *Worker) processNextJob(ctx context.Context) {
	_, _ = w.ProcessSingleJob(ctx)
}
