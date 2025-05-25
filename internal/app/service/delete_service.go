package service

import (
	"context"
	"sync"
	"time"

	"github.com/Roma-F/shortener-url/internal/app/logger"
	"github.com/Roma-F/shortener-url/internal/app/models"
	"github.com/Roma-F/shortener-url/internal/app/repository"
)

const (
	deleteBatchSize     = 100
	deleteFlushInterval = 5 * time.Second
	deleteChannelSize   = 1000
)

type DeleteService struct {
	repo       repository.Repository
	deleteChan chan models.DeleteTask
	batchChan  chan []models.DeleteTask
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewDeleteService(repo repository.Repository) *DeleteService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &DeleteService{
		repo:       repo,
		deleteChan: make(chan models.DeleteTask, deleteChannelSize),
		batchChan:  make(chan []models.DeleteTask, 10),
		ctx:        ctx,
		cancel:     cancel,
	}

	service.startWorkers()

	return service
}

func (ds *DeleteService) startWorkers() {
	ds.wg.Add(1)
	go ds.batchWorker()

	ds.wg.Add(1)
	go ds.deleteWorker()
}

func (ds *DeleteService) batchWorker() {
	defer ds.wg.Done()

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
			case <-ds.ctx.Done():
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

		case <-ds.ctx.Done():
			flush()
			close(ds.batchChan)
			return
		}
	}
}

func (ds *DeleteService) deleteWorker() {
	defer ds.wg.Done()

	for {
		select {
		case batch := <-ds.batchChan:
			if len(batch) > 0 {
				ds.processBatch(batch)
			}
		case <-ds.ctx.Done():
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

func (ds *DeleteService) AddDeleteTask(userID, shortURL string) {
	task := models.DeleteTask{
		UserID:   userID,
		ShortURL: shortURL,
	}

	select {
	case ds.deleteChan <- task:
		logger.Sugar.Debugw("Added delete task", "userID", userID, "shortURL", shortURL)
	default:
		logger.Sugar.Warnw("Delete channel is full, dropping task", "userID", userID, "shortURL", shortURL)
	}
}

func (ds *DeleteService) AddDeleteTasks(userID string, shortURLs []string) {
	for _, shortURL := range shortURLs {
		ds.AddDeleteTask(userID, shortURL)
	}
}

func (ds *DeleteService) Stop() {
	logger.Sugar.Info("Stopping delete service...")
	ds.cancel()
	ds.wg.Wait()
	logger.Sugar.Info("Delete service stopped")
}
