package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Roma-F/shortener-url/internal/app/logger"
	"github.com/Roma-F/shortener-url/internal/app/models"
)

const (
	deleteBatchSize     = 100
	deleteFlushInterval = 5 * time.Second
	deleteChannelSize   = 1000
)

var (
	ErrDeleteChannelFull    = errors.New("delete channel is full")
	ErrPartialDeleteFailure = errors.New("some delete tasks failed")
)

type Repository interface {
	MarkURLsAsDeleted(userID string, shortURLs []string) error
}

type DeleteService struct {
	repo       Repository
	deleteChan chan models.DeleteTask
	batchChan  chan []models.DeleteTask
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewDeleteService(repo Repository) *DeleteService {
	service := &DeleteService{
		repo:       repo,
		deleteChan: make(chan models.DeleteTask, deleteChannelSize),
		batchChan:  make(chan []models.DeleteTask, 10),
	}

	return service
}

func (ds *DeleteService) Start(ctx context.Context) {
	ctx, ds.cancel = context.WithCancel(ctx)
	ds.startWorkers(ctx)
}

func (ds *DeleteService) startWorkers(ctx context.Context) {
	ds.wg.Add(2)

	go func() {
		defer ds.wg.Done()
		ds.batchWorker(ctx)
	}()

	go func() {
		defer ds.wg.Done()
		ds.deleteWorker(ctx)
	}()
}

func (ds *DeleteService) batchWorker(ctx context.Context) {
	batch := make([]models.DeleteTask, 0, deleteBatchSize)
	ticker := time.NewTicker(deleteFlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) > 0 {
			batchCopy := make([]models.DeleteTask, len(batch))
			copy(batchCopy, batch)

			select {
			case ds.batchChan <- batchCopy:
				logger.Sugar.Debugw("Sent delete batch", "size", len(batchCopy))
			case <-ctx.Done():
				return
			}

			batch = batch[:0]
		}
	}

	for {
		select {
		case task := <-ds.deleteChan:
			batch = append(batch, task)

			if len(batch) >= deleteBatchSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-ctx.Done():
			flush()
			close(ds.batchChan)
			return
		}
	}
}

func (ds *DeleteService) deleteWorker(ctx context.Context) {
	for {
		select {
		case batch := <-ds.batchChan:
			if len(batch) > 0 {
				ds.processBatch(batch)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (ds *DeleteService) processBatch(batch []models.DeleteTask) {
	logger.Sugar.Infow("Processing delete batch", "size", len(batch))

	userTasks := make(map[string][]string)
	for _, task := range batch {
		userTasks[task.UserID] = append(userTasks[task.UserID], task.ShortURL)
	}

	for userID, shortURLs := range userTasks {
		err := ds.repo.MarkURLsAsDeleted(userID, shortURLs)
		if err != nil {
			logger.Sugar.Errorw("Failed to mark URLs as deleted",
				"userID", userID,
				"urls", shortURLs,
				"error", err)
		} else {
			logger.Sugar.Infow("Successfully marked URLs as deleted",
				"userID", userID,
				"count", len(shortURLs))
		}
	}
}

func (ds *DeleteService) AddDeleteTask(userID, shortURL string) error {
	task := models.DeleteTask{
		UserID:   userID,
		ShortURL: shortURL,
	}

	select {
	case ds.deleteChan <- task:
		logger.Sugar.Debugw("Added delete task", "userID", userID, "shortURL", shortURL)
		return nil
	default:
		logger.Sugar.Warnw("Delete channel is full, task rejected", "userID", userID, "shortURL", shortURL)
		return ErrDeleteChannelFull
	}
}

func (ds *DeleteService) AddDeleteTasks(userID string, shortURLs []string) error {
	var failedURLs []string

	for _, shortURL := range shortURLs {
		if err := ds.AddDeleteTask(userID, shortURL); err != nil {
			failedURLs = append(failedURLs, shortURL)
		}
	}

	if len(failedURLs) > 0 {
		logger.Sugar.Warnw("Some delete tasks were rejected",
			"userID", userID,
			"failedCount", len(failedURLs),
			"totalCount", len(shortURLs))
		return ErrPartialDeleteFailure
	}

	return nil
}

func (ds *DeleteService) Stop() {
	logger.Sugar.Info("Stopping delete service...")
	if ds.cancel != nil {
		ds.cancel()
	}
	ds.wg.Wait()
	logger.Sugar.Info("Delete service stopped")
}
