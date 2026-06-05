package service

import (
	"github.com/ryoshimaru/MindMap/internal/domain"
	"github.com/ryoshimaru/MindMap/internal/store"
)

type VersionService struct {
	store store.Store
}

func NewVersionService(store store.Store) *VersionService {
	return &VersionService{store: store}
}

func (s *VersionService) List(userID, goalID string) ([]domain.PlanVersion, error) {
	return s.store.ListPlanVersions(userID, goalID)
}

func (s *VersionService) Get(userID, goalID, versionID string) (domain.PlanVersion, error) {
	return s.store.GetPlanVersion(userID, goalID, versionID)
}

func (s *VersionService) Activate(userID, goalID, versionID string) (domain.PlanVersion, error) {
	return s.store.ActivatePlanVersion(userID, goalID, versionID)
}
