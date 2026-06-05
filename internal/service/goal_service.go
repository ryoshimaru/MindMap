package service

import (
	"github.com/ryoshimaru/MindMap/internal/domain"
	"github.com/ryoshimaru/MindMap/internal/store"
)

type GoalService struct {
	store store.Store
}

func NewGoalService(store store.Store) *GoalService {
	return &GoalService{store: store}
}

func (s *GoalService) List(userID string) []domain.Goal {
	return s.store.ListGoals(userID)
}

func (s *GoalService) Create(userID string, req domain.CreateGoalRequest) (domain.Goal, error) {
	return s.store.CreateGoal(userID, req)
}

func (s *GoalService) Get(userID, goalID string) (domain.Goal, error) {
	return s.store.GetGoal(userID, goalID)
}

func (s *GoalService) Update(userID, goalID string, req domain.UpdateGoalRequest) (domain.Goal, error) {
	return s.store.UpdateGoal(userID, goalID, req)
}

func (s *GoalService) Delete(userID, goalID string) error {
	return s.store.DeleteGoal(userID, goalID)
}

func (s *GoalService) GetContext(userID, goalID string) (domain.GoalContext, error) {
	return s.store.GetGoalContext(userID, goalID)
}

func (s *GoalService) CreateContext(userID, goalID string, req domain.UpsertGoalContextRequest) (domain.GoalContext, error) {
	return s.store.CreateGoalContext(userID, goalID, req)
}

func (s *GoalService) UpdateContext(userID, goalID string, req domain.UpsertGoalContextRequest) (domain.GoalContext, error) {
	return s.store.UpdateGoalContext(userID, goalID, req)
}
