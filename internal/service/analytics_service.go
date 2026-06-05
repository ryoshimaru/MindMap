package service

import (
	"github.com/ryoshimaru/MindMap/internal/domain"
	"github.com/ryoshimaru/MindMap/internal/store"
)

type AnalyticsService struct {
	store store.Store
}

func NewAnalyticsService(store store.Store) *AnalyticsService {
	return &AnalyticsService{store: store}
}

func (s *AnalyticsService) Progress(userID, goalID string) (domain.ProgressResponse, error) {
	return s.store.GetProgress(userID, goalID)
}

func (s *AnalyticsService) Feasibility(userID, goalID string) (domain.FeasibilityResponse, error) {
	return s.store.GetFeasibility(userID, goalID)
}

func (s *AnalyticsService) Quality(userID, goalID string) (domain.QualityResponse, error) {
	return s.store.GetQuality(userID, goalID)
}

func (s *AnalyticsService) Metrics(userID, goalID string) (domain.MetricsResponse, error) {
	return s.store.GetMetrics(userID, goalID)
}
