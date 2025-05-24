package service

import (
	"context"
	"fmt"
)

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type HealthService struct {
	db HealthChecker
}

func NewHealthService(db HealthChecker) *HealthService {
	return &HealthService{
		db: db,
	}
}

func (s *HealthService) PingDB(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database not configured")
	}

	return s.db.Ping(ctx)
}
